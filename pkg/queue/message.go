package queue

import (
	"github.com/Azure/moby-packaging/pkg/archive"
)

type Message struct {
	Artifact ArtifactInfo `json:"artifact"`
	Spec     archive.Spec `json:"spec"`
}

type ArtifactInfo struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Sha256Sum string `json:"sha256sum"`
}
