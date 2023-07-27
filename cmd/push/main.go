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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
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
	blobBucketURL        = "https://moby.blob.core.windows.net/"
	containerExistsError = "RESPONSE 409"
	accountName          = "moby"
	queueName            = "moby-packaging-signing-and-publishing"
)

var (
	containerName = fmt.Sprintf("%d", time.Now().Unix())
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

	client, err := azblob.NewClient(blobBucketURL, credential, nil)
	if err != nil {
		panic(err)
	}

	if _, err := client.CreateContainer(ctx, containerName, nil); err != nil {
		if !strings.Contains(err.Error(), containerExistsError) {
			panic(err)
		}
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

		if strings.HasSuffix(d.Name(), "spec.json") {
			specFile = path
		}

		return nil
	}); err != nil {
		panic(err)
	}

	if blobFile == "" {
		panic("no artifact found")
	}

	if specFile == "" {
		panic("no spec file found")
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

	sum, err := getArtifactDigest(blobFile)
	if err != nil {
		panic(err)
	}

	qm := QueueMessage{
		Spec: spec,
		Artifact: ArtifactInfo{
			Name:      filepath.Base(blobFile),
			URI:       fmt.Sprintf("%s%s/%s", blobBucketURL, containerName, storagePathBlob),
			Sha256Sum: sum,
		},
	}

	final, err := json.MarshalIndent(&qm, "", "    ")
	if err != nil {
		panic(err)
	}

	serviceURL := fmt.Sprintf("https://%s.queue.core.windows.net", accountName)

	sCredential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}

	sClient, err := azqueue.NewServiceClient(serviceURL, sCredential, nil)
	if err != nil {
		panic(err)
	}

	qClient := sClient.NewQueueClient(queueName)
	resp, err := qClient.EnqueueMessage(ctx, string(final), &azqueue.EnqueueMessageOptions{TimeToLive: to.Ptr(int32(60) * 60 * 24 * 7)})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(final))
	fmt.Println(resp)
}

func getArtifactDigest(blobFile string) (string, error) {
	b, err := os.ReadFile(blobFile)
	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(b))
	return sha, nil
}
