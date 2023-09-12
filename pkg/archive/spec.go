package archive

import (
	"path/filepath"
	"strings"
)

type Spec struct {
	Pkg      string `json:"package"`
	Distro   string `json:"distro"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo" hash:"ignore"`
	Commit   string `json:"commit" hash:"ignore"`
	Tag      string `json:"tag" hash:"ignore"`
	Revision string `json:"revision" hash:"ignore"`
}

type PipelineInstructions struct {
	Spec                `json:"spec"`
	SpecPath            string `json:"specPath"`
	Basename            string `json:"basename"`
	OriginalArtifactDir string `json:"originalArtifactPath"`
	SignedArtifactDir   string `json:"signedArtifactPath"`
	TestResultsPath     string `json:"testResultsPath"`
	SignedSha256Sum     string `json:"signedSha256Sum"`
}

func (pi *PipelineInstructions) OriginalArtifactPath() string {
	return filepath.Join(pi.OriginalArtifactDir, pi.Basename)
}

func (pi *PipelineInstructions) SignedArtifactPath() string {
	return filepath.Join(pi.SignedArtifactDir, pi.Basename)
}

func (s *Spec) OS() string {
	if s.Distro == "windows" {
		return "windows"
	}

	return "linux"
}

func (s *Spec) SanitizedArch() string {
	return strings.ReplaceAll(s.Arch, "/", "")
}
