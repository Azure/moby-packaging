package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
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
	default:
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	return nil
}

func runDownload(args downloadArgs) error {
	msgs := []azqueue.DequeuedMessage{}

	b, err := os.ReadFile(args.messagesFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &msgs); err != nil {
		return err
	}

	fmt.Printf("%#v\n", msgs)
	return nil
}

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
