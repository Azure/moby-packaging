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
	BuildID     string
	Debug       bool
}

const (
	blobBucketURL        = "https://moby.blob.core.windows.net/"
	containerExistsError = "RESPONSE 409"
	accountName          = "moby"
	queueName            = "moby-packaging-signing-and-publishing"
)

var (
	// containerName            = fmt.Sprintf("%d", time.Now().Unix())
	sevenDaysInSeconds int32 = 60 * 60 * 24 * 7
)

func main() {
	if err := perform(); err != nil {
		panic(err)
	}
}

func perform() error {
	f := Flags{}
	flag.StringVar(&f.ArtifactDir, "artifact-dir", "", "path to directory of artifacts to upload")
	flag.StringVar(&f.BuildID, "build-id", "", "build id")
	flag.BoolVar(&f.Debug, "debug", false, "enable debug output")
	flag.Parse()

	if f.ArtifactDir == "" {
		fmt.Fprintln(os.Stderr, "Arguments must all be provided")
		os.Exit(1)
	}

	ctx := context.Background()
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	client, err := azblob.NewClient(blobBucketURL, credential, nil)
	if err != nil {
		return err
	}

	containerName := f.BuildID
	if containerName == "" {
		containerName = fmt.Sprintf("%d", time.Now().Unix())
	}

	if _, err := client.CreateContainer(ctx, containerName, nil); err != nil {
		if !strings.Contains(err.Error(), containerExistsError) {
			return err
		}
	}

	blobFile, specFile, err := findBlobAndSpec(f)
	if err != nil {
		return err
	}

	spec, err := unmarshalSpec(specFile)
	if err != nil {
		return err
	}

	pkgOS := spec.OS()
	sanitizedArch := strings.ReplaceAll(spec.Arch, "/", "")
	blobBasename := filepath.Base(blobFile)
	specBasename := filepath.Base(specFile)
	storagePathBlob := fmt.Sprintf("%s/%s+azure/%s/%s_%s/%s", spec.Pkg, spec.Tag, spec.Distro, pkgOS, sanitizedArch, blobBasename)
	storagePathSpec := fmt.Sprintf("%s/%s+azure/%s/%s_%s/%s", spec.Pkg, spec.Tag, spec.Distro, pkgOS, sanitizedArch, specBasename)

	blobGoFile, err := os.Open(blobFile)
	if err != nil {
		return err
	}

	specGoFile, err := os.Open(specFile)
	if err != nil {
		return err
	}

	if _, err := client.UploadFile(ctx, containerName, storagePathBlob, blobGoFile, &azblob.UploadFileOptions{}); err != nil {
		return err
	}
	fmt.Println("file uploaded:", storagePathBlob)

	if _, err := client.UploadFile(ctx, containerName, storagePathSpec, specGoFile, &azblob.UploadFileOptions{}); err != nil {
		return err
	}
	fmt.Println("file uploaded:", storagePathSpec)

	sum, err := getArtifactDigest(blobFile)
	if err != nil {
		return err
	}

	qm := QueueMessage{
		Spec: spec,
		Artifact: ArtifactInfo{
			Name:      filepath.Base(blobFile),
			URI:       fmt.Sprintf("%s%s/%s", blobBucketURL, containerName, storagePathBlob),
			Sha256Sum: sum,
		},
	}

	queueMessageCompact, err := json.Marshal(&qm)
	if err != nil {
		return err
	}

	serviceURL := fmt.Sprintf("https://%s.queue.core.windows.net", accountName)

	sClient, err := azqueue.NewServiceClient(serviceURL, credential, nil)
	if err != nil {
		return err
	}

	qClient := sClient.NewQueueClient(queueName)

	resp, err := qClient.EnqueueMessage(ctx, string(queueMessageCompact), &azqueue.EnqueueMessageOptions{TimeToLive: &sevenDaysInSeconds})
	if err != nil {
		return err
	}

	if !f.Debug {
		return nil
	}

	// debug output
	queueMessageHuman, err := json.MarshalIndent(&qm, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(queueMessageHuman)
	fmt.Printf("%#v\n", resp)

	return nil
}

func unmarshalSpec(specFile string) (archive.Spec, error) {
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return archive.Spec{}, err
	}

	var spec archive.Spec
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return archive.Spec{}, err
	}
	return spec, nil
}

func findBlobAndSpec(f Flags) (string, string, error) {
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
		return "", "", err
	}

	if blobFile == "" || specFile == "" {
		return "", "", fmt.Errorf("blob file and spec file must be present in artifact dir\nblob: %s\nspec:%s\n", blobFile, specFile)
	}

	return blobFile, specFile, nil
}

func getArtifactDigest(blobFile string) (string, error) {
	b, err := os.ReadFile(blobFile)
	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(b))
	return sha, nil
}
