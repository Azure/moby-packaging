package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/moby-packaging/pkg/archive"
)

type QueueMessage struct {
	Artifact ArtifactInfo `json:"artifact"`
	Spec     archive.Spec `json:"spec"`
}

type ArtifactInfo struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Sha256Sum string `json:"sha256sum"`
}

type Flags struct {
	ArtifactDir string
}

type BlobsJSON []BlobKV

type BlobKV struct {
	Blob string
}

const (
	url           = "https://moby.blob.core.windows.net/"
	containerName = "moby"
)

func main() {
	f := Flags{}
	flag.StringVar(&f.ArtifactDir, "artifact-dir", "", "path to directory of artifacts to upload")
	flag.Parse()

	if f.ArtifactDir == "" {
		fmt.Fprintln(os.Stderr, "Arguments must all be provided")
		os.Exit(1)
	}

	ctx := context.Background()
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	client, err := azblob.NewClient(url, credential, nil)
	if err != nil {
		panic(err)
	}

	if _, err := client.CreateContainer(ctx, containerName, nil); err != nil {
		panic(err)
	}

	var blobFile string
	var specFile string
	r := regexp.MustCompile(`^.*\.(deb|rpm|zip)$`)
	if err := filepath.WalkDir(f.ArtifactDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if r.MatchString(d.Name()) {
			blobFile = path
			return nil
		}

		if strings.HasSuffix(specFile, "spec.json") {
			specFile = path
		}

		return nil
	}); err != nil {
		panic(err)
	}

	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		panic(err)
	}

	var spec archive.Spec
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		panic(err)
	}

	pkgOS := "linux"
	if spec.Distro == "windows" {
		pkgOS = "windows"
	}

	sanitizedArch := strings.ReplaceAll(spec.Arch, "/", "")
	blobBasename := filepath.Base(blobFile)
	specBasename := filepath.Base(specFile)
	storagePathBlob := fmt.Sprintf("%s/%s+azure/%s/%s_%s/%s", spec.Pkg, spec.Tag, spec.Distro, pkgOS, sanitizedArch, blobBasename)
	storagePathSpec := fmt.Sprintf("%s/%s+azure/%s/%s_%s/%s", spec.Pkg, spec.Tag, spec.Distro, pkgOS, sanitizedArch, specBasename)

	blob, err := os.Open(blobFile)
	if err != nil {
		panic(err)
	}

	specGoFile, err := os.Open(specFile)
	if err != nil {
		panic(err)
	}

	if _, err := client.UploadFile(ctx, containerName, storagePathBlob, blob, &azblob.UploadFileOptions{}); err != nil {
		panic(err)
	}

	if _, err := client.UploadFile(ctx, containerName, storagePathSpec, specGoFile, &azblob.UploadFileOptions{}); err != nil {
		panic(err)
	}

	qm := QueueMessage{}
	qm.Spec = spec
	qm.Artifact.Name = filepath.Base(f.ArtifactDir)

	qm.Artifact.Sha256Sum, err = getArtifactDigest(f)
	if err != nil {
		panic(err)
	}
	artifactURI := fmt.Sprintf("%s%s/%s", url, containerName, storagePathBlob)
	qm.Artifact.URI = artifactURI

	final, err := json.MarshalIndent(&qm, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(final))
}

func getArtifactDigest(f Flags) (string, error) {
	b, err := os.ReadFile(f.ArtifactDir)
	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(b))
	return sha, nil
}
