package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Azure/moby-packaging/pkg/queue"
)

const (
	accountName = "moby"
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
	inDir        string
	messagesFile string
}

type deleteArgs struct {
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

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	dlArgs := downloadArgs{}
	upArgs := uploadArgs{}
	dtArgs := deleteArgs{}
	var messagesFile string

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&dlArgs.outDir, "out-dir", "", "directory to download files")
	fs.StringVar(&upArgs.inDir, "in-dir", "", "directory containing files to upload")
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
	case "delete":
		if err := runDelete(dtArgs); err != nil {
			return err
		}
	case "fetch": // fetch messages
		if err := runFetch(); err != nil {
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

	msgs := []QueueMessageDeserialize{}
	downloaded := []QueueMessageDeserialize{}

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

		resp, err := c.Get(uri)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		blobContents := new(bytes.Buffer)
		if _, err := io.Copy(blobContents, resp.Body); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		b := blobContents.Bytes()

		actualSum := fmt.Sprintf("%x", sha256.Sum256(b))
		if actualSum != expectedSum {
			err := fmt.Errorf("wrong sum for artifact %s\n\texpected: %s\n\tactual:%s", msg.Content.Artifact.Name, expectedSum, actualSum)
			errs = errors.Join(errs, err)
			continue
		}

		dstFile := filepath.Join(args.outDir, msg.Content.Artifact.Name)
		if err := os.WriteFile(dstFile, b, 0o600); err != nil {
			errs = errors.Join(errs, err)
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

	fmt.Println(string(s))
	return nil
}

// msgContents := []queue.Message{}
// for _, msg := range msgs {
// 	msgContents = append(msgContents, msg.Content)
// }

//

// x, err := json.MarshalIndent(msgContents, "", "    ")
// fmt.Println(string(x))

func runUpload(args uploadArgs) error {
	_ = args

	fmt.Println(args)
	return nil
}

func runDelete(args deleteArgs) error {
	_ = args

	fmt.Println(args)
	return nil
}

func runFetch() error {

	return nil
}
