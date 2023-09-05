package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/Azure/moby-packaging/pkg/queue"
)

const (
	stagingAccountName = "moby"
	prodAccountName    = "mobyreleases"
	prodContainerName  = "moby"
	queueName          = "moby-packaging-signing-and-publishing"

	sha256Key = "sha256"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", err)
		os.Exit(1)
	}
}

type Envelope struct {
	Content         string `json:"content"`
	DequeueCount    int    `json:"dequeueCount"`
	ExpirationTime  string `json:"expirationTime"`
	ID              string `json:"id"`
	InsertionTime   string `json:"insertionTime"`
	PopReceipt      string `json:"popReceipt"`
	TimeNextVisible string `json:"timeNextVisible"`
}

func (e *Envelope) GetMessageContent() (queue.Message, error) {
	b, err := base64.StdEncoding.DecodeString(e.Content)
	if err != nil {
		return queue.Message{}, err
	}

	var msg queue.Message
	if err := json.Unmarshal(b, &msg); err != nil {
		return queue.Message{}, err
	}

	return msg, nil
}

type downloadArgs struct {
	outDir       string
	messagesFile string
}

type uploadArgs struct {
	signedDir    string
	messagesFile string
}

type fixupQueueArgs struct {
	messagesFile string
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("available arguments are get-messages, download, upload, and delete")
	}

	dlArgs := downloadArgs{}
	upArgs := uploadArgs{}
	fqArgs := fixupQueueArgs{}
	var messagesFile string

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&dlArgs.outDir, "out-dir", "", "directory to download files")
	fs.StringVar(&upArgs.signedDir, "signed-dir", "", "directory containing files to upload")
	fs.StringVar(&messagesFile, "messages-file", "", "file containing queue messages to process")
	fs.Parse(os.Args[2:])

	dlArgs.messagesFile = messagesFile
	upArgs.messagesFile = messagesFile
	fqArgs.messagesFile = messagesFile

	switch os.Args[1] {
	case "get-messages":
		if err := getMessages(); err != nil {
			return err
		}
	case "download":
		if err := runDownload(dlArgs); err != nil {
			return err
		}
	case "upload":
		if err := runUpload(upArgs); err != nil {
			return err
		}
	case "fixup-queue":
		if err := runFixupQueue(fqArgs); err != nil {
			return err
		}
	default:
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	return nil
}

func getMessages() error {
	ctx := context.Background()
	q, err := queue.NewDefaultSignQueueClient()
	if err != nil {
		return err
	}

	msgs, err := q.GetAllMessages(ctx)
	if err != nil {
		if len(msgs.Messages) == 0 {
			return err
		}
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]errors getting messages: %s\n", err)
	}

	envelopes, err := convertToRawMsgs(msgs.Messages)
	if err != nil {
		if len(envelopes) == 0 {
			return err
		}
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]errors converting messages to raw messages: %s\n", err)
	}

	j, err := json.Marshal(&envelopes)
	if err != nil {
		return err
	}

	// Write to stdout
	fmt.Println(string(j))

	return nil
}

func convertToRawMsgs(msgs []*azqueue.DequeuedMessage) ([]Envelope, error) {
	envelopes := []Envelope{}
	var errs error

	fail := func(collection *error, err error) {
		*collection = errors.Join(*collection, err)
	}

	for _, msg := range msgs {
		var e Envelope
		if msg == nil {
			fail(&errs, fmt.Errorf("nil message, skipping..."))
			continue
		}

		if msg.MessageID == nil {
			fail(&errs, fmt.Errorf("nil message ID, skipping..."))
			continue
		}
		e.ID = *msg.MessageID

		if msg.MessageText == nil {
			fail(&errs, fmt.Errorf("nil message content, skipping..."))
			continue
		}
		e.Content = *msg.MessageText

		if msg.PopReceipt == nil {
			fail(&errs, fmt.Errorf("nil message popReceipt, skipping"))
			continue
		}
		e.PopReceipt = *msg.PopReceipt

		if msg.DequeueCount != nil {
			e.DequeueCount = int(*msg.DequeueCount)
		}

		if msg.ExpirationTime != nil {
			e.ExpirationTime = msg.ExpirationTime.Format(time.RFC3339)
		}

		if msg.InsertionTime != nil {
			e.InsertionTime = msg.InsertionTime.Format(time.RFC3339)
		}

		if msg.TimeNextVisible != nil {
			e.TimeNextVisible = msg.TimeNextVisible.Format(time.RFC3339)
		}

		envelopes = append(envelopes, e)
	}

	return envelopes, errs
}

func runDownload(args downloadArgs) error {
	if args.outDir == "" {
		return fmt.Errorf("you must specify a directory to download into")
	}

	if err := os.MkdirAll(args.outDir, 0o700); err != nil {
		return err
	}

	envelopes := []Envelope{}
	downloaded := []Envelope{}
	failed := []Envelope{}

	b, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &envelopes); err != nil {
		return err
	}

	c := http.Client{}

	var errs error
	fail := func(localEnvelope Envelope, localErr error) {
		failed = append(failed, localEnvelope)
		errs = errors.Join(errs, localErr)
	}

	// Loop over messages, downloading blobs for each
	for _, envelope := range envelopes {
		// if there's a failure during the downloading process, do not add the msg to the downloaded array
		message, err := envelope.GetMessageContent()
		if err != nil {
			fail(envelope, err)
			continue
		}
		uri := message.Artifact.URI
		expectedSum := message.Artifact.Sha256Sum

		resp, err := c.Get(uri)
		if err != nil {
			fail(envelope, err)
			continue
		}

		blobContents := new(bytes.Buffer)
		if _, err := io.Copy(blobContents, resp.Body); err != nil {
			fail(envelope, err)
			continue
		}

		b := blobContents.Bytes()

		actualSum := fmt.Sprintf("%x", sha256.Sum256(b))
		if actualSum != expectedSum {
			fail(envelope, fmt.Errorf("wrong sum for artifact %s\n\texpected: %s\n\tactual:%s", message.Artifact.Name, expectedSum, actualSum))
			continue
		}

		dstFile := filepath.Join(args.outDir, message.Artifact.Name)
		if err := os.WriteFile(dstFile, b, 0o600); err != nil {
			fail(envelope, err)
			continue
		}

		downloaded = append(downloaded, envelope)
	}

	if errs != nil {
		errStrs := strings.Split(errs.Error(), "\n")
		for _, e := range errStrs {
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", e)
		}
	}

	// After completion, print the downloaded array to stdout as JSON
	s, err := json.MarshalIndent(&downloaded, "", "    ")
	if err != nil {
		return err
	}

	msgDir := filepath.Dir(args.messagesFile)
	failedFile := filepath.Join(msgDir, "failed_downloading")
	failedJSON, err := json.MarshalIndent(&failed, "", "    ")
	if err != nil {
		return err
	}

	// This is not a failure condition, since the failed ones may be retried
	_ = os.WriteFile(failedFile, failedJSON, 0o600)

	fmt.Println(string(s))
	return nil
}

func runUpload(args uploadArgs) error {
	if args.messagesFile == "" {
		return fmt.Errorf("you must provide a messages file")
	}

	if args.signedDir == "" {
		return fmt.Errorf("you must provide a directory for the signed packages")
	}

	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	msgBytes, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	allEnvelopes := []Envelope{}
	failed := []Envelope{}
	successful := []Envelope{}
	var errs error

	if err := json.Unmarshal(msgBytes, &allEnvelopes); err != nil {
		return err
	}

	fail := func(localErr error, localEnvelope ...Envelope) {
		for _, c := range localEnvelope {
			failed = append(failed, c)
		}
		errs = errors.Join(errs, localErr)
	}

	nameToEnvelopes := make(map[string][]Envelope)
	for _, envelope := range allEnvelopes {
		message, err := envelope.GetMessageContent()
		if err != nil {
			fail(err, envelope)
			continue
		}

		nameToEnvelopes[message.Artifact.Name] = append(nameToEnvelopes[message.Artifact.Name], envelope)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", prodAccountName)
	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		return err
	}

	for pkgBasename, pkgEnvelopes := range nameToEnvelopes {
		signedPkgFilename := filepath.Join(args.signedDir, pkgBasename)
		if _, err := os.Stat(signedPkgFilename); err != nil {
			err := err
			fail(err, pkgEnvelopes...)
			continue
		}

		envelope, err := resolveDuplicates(pkgEnvelopes)
		if err != nil {
			fail(err, pkgEnvelopes...)
			continue
		}

		message, err := envelope.GetMessageContent()
		if err != nil {
			fail(err, pkgEnvelopes...)
			continue
		}

		pkg := message.Spec.Pkg
		version := fmt.Sprintf("%s+azure", message.Spec.Tag)
		distro := message.Spec.Distro
		pkgOS := message.Spec.OS()
		sanitizedArch := strings.ReplaceAll(message.Spec.Arch, "/", "_")

		storagePath := fmt.Sprintf("%s/%s/%s/%s_%s/%s", pkg, version, distro, pkgOS, sanitizedArch, pkgBasename)

		b, err := os.ReadFile(signedPkgFilename)
		if err != nil {
			fail(err, pkgEnvelopes...)
			continue
		}

		signedSha256Sum := fmt.Sprintf("%x", sha256.Sum256(b))

		if _, err := client.UploadBuffer(ctx, prodContainerName, storagePath, b, &azblob.UploadFileOptions{
			Metadata: map[string]*string{sha256Key: &signedSha256Sum},
		}); err != nil {
			fail(err, pkgEnvelopes...)
			continue
		}

		successful = append(successful, pkgEnvelopes...)
	}

	if errs != nil {
		errStrs := strings.Split(errs.Error(), "\n")
		for _, e := range errStrs {
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", e)
		}
	}

	// After completion, print the downloaded array to stdout as JSON
	s, err := json.MarshalIndent(&successful, "", "    ")
	if err != nil {
		return err
	}

	msgDir := filepath.Dir(args.messagesFile)
	failedFile := filepath.Join(msgDir, "failed_singing_or_publishing")
	failedJSON, err := json.MarshalIndent(&failed, "", "    ")
	if err != nil {
		return err
	}

	// This is not a failure condition, since the failed ones may be retried
	_ = os.WriteFile(failedFile, failedJSON, 0o600)

	fmt.Println(string(s))
	return nil
}

func runFixupQueue(args fixupQueueArgs) error {
	if args.messagesFile == "" {
		return fmt.Errorf("you must provide a messages file")
	}

	ctx := context.Background()
	_ = ctx
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	_ = cred

	if err != nil {
		return err
	}

	serviceURL := fmt.Sprintf("https://%s.queue.core.windows.net", stagingAccountName)

	sClient, err := azqueue.NewServiceClient(serviceURL, cred, nil)
	if err != nil {
		return err
	}

	qClient := sClient.NewQueueClient(queueName)

	msgBytes, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	envelopes := []Envelope{}
	if err := json.Unmarshal(msgBytes, &envelopes); err != nil {
		return err
	}

	failed := []azqueue.DeleteMessageResponse{}
	succeeded := []azqueue.DeleteMessageResponse{}

	var errs error
	for _, envelope := range envelopes {
		resp, err := qClient.DeleteMessage(ctx, envelope.ID, envelope.PopReceipt, &azqueue.DeleteMessageOptions{})
		if err != nil {
			errs = errors.Join(errs, err)
			failed = append(failed, resp)
			continue
		}

		succeeded = append(succeeded, resp)
	}

	if errs != nil {
		all := strings.Split(errs.Error(), "\n")
		for _, err := range all {
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", err)
		}

	}

	msgDir := filepath.Dir(args.messagesFile)
	failedFile := filepath.Join(msgDir, "failed_deleting_from_queue")
	failedJSON, err := json.MarshalIndent(&failed, "", "    ")
	if err != nil {
		return err
	}

	// This is not a failure condition, since the failed ones may be retried
	_ = os.WriteFile(failedFile, failedJSON, 0o600)

	return nil
}

func resolveDuplicates(e []Envelope) (Envelope, error) {
	switch len(e) {
	case 0:
		err := fmt.Errorf("unexpected error: the length of the envelopes array is zero")
		return Envelope{}, err
	case 1:
		return e[0], nil
	default:
		lastMsg, err := e[0].GetMessageContent()
		if err != nil {
			return Envelope{}, err
		}

		for i := 1; i < len(e); i++ {
			thisMsg, err := e[i].GetMessageContent()
			if err != nil {
				return Envelope{}, err
			}

			if lastMsg.Artifact.Sha256Sum != thisMsg.Artifact.Sha256Sum {
				return Envelope{}, fmt.Errorf(
					"messages encountered with same filename and different sha256 digests; manual intervention will be required. "+
						"filename: `%s`, digest[%d]: %s digest[%d]: %s",
					lastMsg.Artifact.Name, i-1, lastMsg.Artifact.Sha256Sum, i, thisMsg.Artifact.Sha256Sum)
			}

			lastMsg = thisMsg
		}

		return e[0], nil
	}
}
