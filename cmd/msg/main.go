package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	BlobsFile        string
	ArtifactFilename string
	SpecFilename     string
}

type BlobsJSON []BlobKV

type BlobKV struct {
	Blob string `json:"blob"`
}

func main() {
	f := Flags{}
	flag.StringVar(&f.BlobsFile, "blobs-file", "", "path of the blobs.json file to read")
	flag.StringVar(&f.ArtifactFilename, "artifact-filename", "", "filename of the artifact to enqueue")
	flag.StringVar(&f.SpecFilename, "spec-filename", "", "filename of the spec to enqueue")
	flag.Parse()

	if f.BlobsFile == "" || f.ArtifactFilename == "" || f.SpecFilename == "" {
		fmt.Fprintln(os.Stderr, "Arguments must all be provided")
		os.Exit(1)
	}

	qm := QueueMessage{}
	qm.Artifact.Name = filepath.Base(f.ArtifactFilename)

	blobs := getBlobInfo(f)
	artifactURI := getArtifactURI(blobs)

	var err error
	qm.Artifact.Sha256Sum, err = getArtifactDigest(f)
	if err != nil {
		panic(err)
	}
	qm.Artifact.URI = artifactURI

	specBytes, err := os.ReadFile(f.SpecFilename)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(specBytes, &qm.Spec); err != nil {
		panic(err)
	}

	final, err := json.MarshalIndent(&qm, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(final))
}

func getArtifactDigest(f Flags) (string, error) {
	b, err := os.ReadFile(f.ArtifactFilename)
	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", sha256.Sum256(b))
	return sha, nil
}

func getArtifactURI(blobs BlobsJSON) string {
	idx := 0
	if strings.HasSuffix(blobs[0].Blob, "spec.json") {
		idx = 1
	}
	artifactURI := blobs[idx].Blob
	return artifactURI
}

func getBlobInfo(f Flags) BlobsJSON {
	blobs := BlobsJSON{}
	blobsBytes, err := os.ReadFile(f.BlobsFile)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(blobsBytes, &blobs); err != nil {
		panic(err)
	}

	if len(blobs) != 2 {
		fmt.Fprintln(os.Stderr, "There should be only two files uploaded: the artifact and the spec file")
		os.Exit(1)
	}
	return blobs
}
