package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
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

type QueueMessageDeserialize struct {
	Content         queue.Message `json:"content"`
	DequeueCount    int           `json:"dequeueCount"`
	ExpirationTime  string        `json:"expirationTime"`
	ID              string        `json:"id"`
	InsertionTime   string        `json:"insertionTime"`
	PopReceipt      string        `json:"popReceipt"`
	TimeNextVisible string        `json:"timeNextVisible"`
}

func (m *QueueMessageDeserialize) UnmarshalJSON(data []byte) error {
	// we have to do this to keep from exploding the call stack
	// see - https://medium.com/@turgon/json-in-go-is-magical-c5b71505a937

	type Aux struct {
		Content         string `json:"content"`
		DequeueCount    int    `json:"dequeueCount"`
		ExpirationTime  string `json:"expirationTime"`
		ID              string `json:"id"`
		InsertionTime   string `json:"insertionTime"`
		PopReceipt      string `json:"popReceipt"`
		TimeNextVisible string `json:"timeNextVisible"`
	}

	var aux Aux
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.DequeueCount = aux.DequeueCount
	m.ExpirationTime = aux.ExpirationTime
	m.ID = aux.ID
	m.InsertionTime = aux.InsertionTime
	m.PopReceipt = aux.PopReceipt
	m.TimeNextVisible = aux.TimeNextVisible

	// return json.Unmarshal(aux.Content, &m.Content)
	if err := json.Unmarshal([]byte(aux.Content), &m.Content); err != nil {
		return err
	}

	return nil
}

func (m *QueueMessageDeserialize) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(m.Content)
	if err != nil {
		return nil, err
	}

	type Aux struct {
		Content         string `json:"content"`
		DequeueCount    int    `json:"dequeueCount"`
		ExpirationTime  string `json:"expirationTime"`
		ID              string `json:"id"`
		InsertionTime   string `json:"insertionTime"`
		PopReceipt      string `json:"popReceipt"`
		TimeNextVisible string `json:"timeNextVisible"`
	}

	aux := Aux{
		Content:         string(b),
		DequeueCount:    m.DequeueCount,
		ExpirationTime:  m.ExpirationTime,
		ID:              m.ID,
		InsertionTime:   m.InsertionTime,
		PopReceipt:      m.PopReceipt,
		TimeNextVisible: m.TimeNextVisible,
	}

	return json.Marshal(aux)
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	dlArgs := downloadArgs{}
	upArgs := uploadArgs{}
	dtArgs := fixupQueueArgs{}
	var messagesFile string

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&dlArgs.outDir, "out-dir", "", "directory to download files")
	fs.StringVar(&upArgs.signedDir, "signed-dir", "", "directory containing files to upload")
	fs.StringVar(&messagesFile, "messages-file", "", "file containing queue messages to process")
	fs.Parse(os.Args[2:])

	dlArgs.messagesFile = messagesFile
	upArgs.messagesFile = messagesFile
	dtArgs.messagesFile = messagesFile

	switch os.Args[1] {
	case "download":
		if err := runDownload(dlArgs); err != nil {
			return err
		}
	case "upload":
		if err := runUpload(upArgs); err != nil {
			return err
		}
	case "fixup-queue":
		if err := runFixupQueue(dtArgs); err != nil {
			return err
		}
	default:
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	return nil
}

func runDownload(args downloadArgs) error {
	if args.outDir == "" {
		return fmt.Errorf("you must specify a directory to download into")
	}

	if err := os.MkdirAll(args.outDir, 0o700); err != nil {
		return err
	}

	msgs := []QueueMessageDeserialize{}
	downloaded := []QueueMessageDeserialize{}
	failed := []QueueMessageDeserialize{}

	b, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &msgs); err != nil {
		return err
	}

	// Loop over messages, downloading blobs for each
	var errs error
	c := http.Client{}
	for _, msg := range msgs {
		// if there's a failure during the downloading process, do not add the msg to the downloaded array
		uri := msg.Content.Artifact.URI
		expectedSum := msg.Content.Artifact.Sha256Sum

		fail := func(f QueueMessageDeserialize, e error) {
			failed = append(failed, f)
			errs = errors.Join(errs, err)
		}

		resp, err := c.Get(uri)
		if err != nil {
			fail(msg, err)
			continue
		}

		blobContents := new(bytes.Buffer)
		if _, err := io.Copy(blobContents, resp.Body); err != nil {
			fail(msg, err)
			continue
		}

		b := blobContents.Bytes()

		actualSum := fmt.Sprintf("%x", sha256.Sum256(b))
		if actualSum != expectedSum {
			fail(msg, fmt.Errorf("wrong sum for artifact %s\n\texpected: %s\n\tactual:%s", msg.Content.Artifact.Name, expectedSum, actualSum))
			continue
		}

		dstFile := filepath.Join(args.outDir, msg.Content.Artifact.Name)
		if err := os.WriteFile(dstFile, b, 0o600); err != nil {
			fail(msg, err)
			continue
		}

		downloaded = append(downloaded, msg)
	}

	if errs != nil {
		fmt.Fprintln(os.Stderr, errs)
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
	_ = ctx
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	msgBytes, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	msgs := []QueueMessageDeserialize{}
	if err := json.Unmarshal(msgBytes, &msgs); err != nil {
		return err
	}

	var errs error
	nameToMsg := make(map[string][]QueueMessageDeserialize)
	for _, msg := range msgs {
		nameToMsg[msg.Content.Artifact.Name] = append(nameToMsg[msg.Content.Artifact.Name], msg)
	}

	failed := []QueueMessageDeserialize{}
	successful := []QueueMessageDeserialize{}

	fail := func(e error, f ...QueueMessageDeserialize) {
		for _, c := range f {
			failed = append(failed, c)
		}
		errs = errors.Join(errs, err)
	}

	client, err := azblob.NewClient(prodAccountName, cred, nil)
	_ = client
	if err != nil {
		return err
	}

outer:
	for pkgBasename, messages := range nameToMsg {
		signedPkgFilename := filepath.Join(args.signedDir, pkgBasename)
		if _, err := os.Stat(signedPkgFilename); err != nil {
			errs = errors.Join(errs, fmt.Errorf("filename not found in the signing directory; signing likely failed for '%s'", pkgBasename))
			continue
		}

		if len(messages) > 1 {
			for i := 1; i < len(messages); i++ {
				if messages[i-1].Content.Artifact.Sha256Sum != messages[i].Content.Artifact.Sha256Sum {
					fail(
						fmt.Errorf(
							"messages encountered with same filename and different sha256 digests; manual intervention will be required. "+
								"digest[%d]: %s digest[%d]: %s",
							i-1, messages[i-1].Content.Artifact.Sha256Sum, i, messages[i].Content.Artifact.Sha256Sum),
						messages...,
					)
					continue outer
				}
			}
		}

		msg := messages[0]

		pkg := msg.Content.Spec.Pkg
		version := fmt.Sprintf("%s+azure", msg.Content.Spec.Tag)
		distro := msg.Content.Spec.Distro
		pkgOS := msg.Content.Spec.OS()
		sanitizedArch := strings.ReplaceAll(msg.Content.Spec.Arch, "/", "")

		storagePath := fmt.Sprintf("%s/%s/%s/%s_%s/%s", pkg, version, distro, pkgOS, sanitizedArch, pkgBasename)
		fmt.Fprintln(os.Stderr, storagePath)

		f, err := os.Open(signedPkgFilename)
		if err != nil {
			fail(err, msg)
			continue
		}

		_ = f

		// if _, err := client.UploadFile(ctx, prodContainerName, storagePath, f, &azblob.UploadFileOptions{}); err != nil {
		// 	fail(err, msg)
		// 	continue
		// }

		successful = append(successful, messages...)
	}

	if errs != nil {
		fmt.Fprintln(os.Stderr, errs)
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

	messages := []QueueMessageDeserialize{}
	if err := json.Unmarshal(msgBytes, &messages); err != nil {
		return err
	}

	failed := []azqueue.DeleteMessageResponse{}
	succeeded := []azqueue.DeleteMessageResponse{}

	var errs error
	for _, msg := range messages {
		resp, err := qClient.DeleteMessage(ctx, msg.ID, msg.PopReceipt, &azqueue.DeleteMessageOptions{})
		if err != nil {
			errs = errors.Join(errs, err)
			failed = append(failed, resp)
			continue
		}

		succeeded = append(succeeded, resp)
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
