package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/queue"
)

const (
	prodAccountName   = "mobyreleases"
	prodContainerName = "moby"

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

type uploadArgs struct {
	signedDir string
	specsFile string
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("available arguments are get-messages, download, upload, and delete")
	}

	upArgs := uploadArgs{}
	var messagesFile string

	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.StringVar(&upArgs.signedDir, "signed-dir", "", "directory containing signed files to upload")
	fs.StringVar(&messagesFile, "specs-file", "", "file containing build specs of files to upload")
	fs.Parse(os.Args[2:])

	upArgs.specsFile = messagesFile

	switch os.Args[1] {
	case "upload":
		if err := runUpload(upArgs); err != nil {
			return err
		}
	default:
		return fmt.Errorf("available arguments are download, upload, and delete")
	}

	return nil
}

func runUpload(args uploadArgs) error {
	if args.specsFile == "" {
		return fmt.Errorf("you must provide a spec file")
	}

	if args.signedDir == "" {
		return fmt.Errorf("you must provide a directory for the signed packages")
	}

	ctx := context.Background()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	specsBytes, err := os.ReadFile(args.specsFile)
	if err != nil {
		return err
	}

	allSpecs := []archive.Spec{}
	if err := json.Unmarshal(specsBytes, &allSpecs); err != nil {
		return err
	}

	failed := make([]archive.Spec, 0, len(allSpecs))
	successful := make([]archive.Spec, 0, len(allSpecs))
	errs := make([]error, 0, len(allSpecs))

	fail := func(e error, s archive.Spec) {
		failed = append(failed, s)
		errs = append(errs, e)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", prodAccountName)
	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		return err
	}

	for _, spec := range allSpecs {
		pkgOS := spec.OS()
		sanitizedArch := strings.ReplaceAll(spec.Arch, "/", "_")
		osArchDir := fmt.Sprintf("%s_%s", pkgOS, sanitizedArch)
		signedPkgDir := filepath.Join(args.signedDir, spec.Distro, osArchDir)
		signedPkgGlob := filepath.Join(signedPkgDir, "*")

		files, err := filepath.Glob(signedPkgGlob)
		if err != nil {
			fail(err, spec)
			continue
		}

		if len(files) != 1 {
			err := fmt.Errorf("Zero or multiple files found matching glob: '%s'", signedPkgGlob)
			fail(err, spec)
			continue
		}

		signedPkgPath := files[0]
		base := filepath.Base(signedPkgPath)
		pkg := spec.Pkg
		version := fmt.Sprintf("%s+azure", spec.Tag)
		distro := spec.Distro
		storagePath := fmt.Sprintf("%s/%s/%s/%s_%s/%s", pkg, version, distro, pkgOS, sanitizedArch, base)

		b, err := os.ReadFile(signedPkgPath)
		if err != nil {
			fail(err, spec)
			continue
		}

		signedSha256Sum := fmt.Sprintf("%x", sha256.Sum256(b))
		if _, err := client.UploadBuffer(ctx, prodContainerName, storagePath, b, &azblob.UploadFileOptions{
			Metadata: map[string]*string{sha256Key: &signedSha256Sum},
		}); err != nil {
			fail(err, spec)
			continue
		}

		successful = append(successful, spec)
	}

	if len(errs) != 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", e)
		}
	}

	// After completion, print the downloaded array to stdout as JSON
	s, err := json.MarshalIndent(&successful, "", "    ")
	if err != nil {
		return err
	}

	for _, f := range failed {
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s %s-%s for %s/%s failed to upload\n", f.Pkg, f.Tag, f.Revision, f.Distro, f.Arch)
	}

	fmt.Println(string(s))
	return nil
}
