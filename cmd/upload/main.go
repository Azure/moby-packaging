package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/moby-packaging/pkg/archive"
)

const (
	prodAccountName   = "mobyreleases"
	prodContainerName = "moby"

	sha256Key = "sha256"
)

type uploadArgs struct {
	signedDir string
	specsFile string
}

func main() {
	upArgs := uploadArgs{}
	flag.StringVar(&upArgs.signedDir, "signed-dir", "", "directory containing signed files to upload")
	flag.StringVar(&upArgs.specsFile, "specs-file", "", "file containing build specs of files to upload")
	flag.Parse()

	if err := do(upArgs); err != nil {
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", err)
		os.Exit(1)
	}
}

func do(args uploadArgs) error {
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

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", prodAccountName)
	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		return err
	}

	failed := make([]archive.Spec, 0, len(allSpecs))
	successful := make([]archive.Spec, 0, len(allSpecs))
	errs := make([]error, 0, len(allSpecs))

	fail := func(e error, s archive.Spec) {
		failed = append(failed, s)
		errs = append(errs, e)
	}

	for _, spec := range allSpecs {
		signedPkgPath, err := spec.FullPath(args.signedDir)
		if err != nil {
			fail(err, spec)
			continue
		}

		storagePath, err := spec.StoragePath()
		if err != nil {
			fail(err, spec)
			continue
		}

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

	for _, f := range failed {
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s %s-%s for %s/%s failed to upload\n", f.Pkg, f.Tag, f.Revision, f.Distro, f.Arch)
	}

	// After completion, print the downloaded array to stdout as JSON
	s, err := json.MarshalIndent(&successful, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(s))
	return nil
}
