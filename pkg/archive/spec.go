package archive

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Spec struct {
	Pkg      string `json:"package"`
	Distro   string `json:"distro"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	Commit   string `json:"commit"`
	Tag      string `json:"tag"`
	Revision string `json:"revision"`
}

type HashOptions struct {
	Pkg      bool
	Distro   bool
	Arch     bool
	Repo     bool
	Commit   bool
	Tag      bool
	Revision bool
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

func (s *Spec) Hash(o HashOptions) (string, error) {
	format := ""
	vals := []interface{}{}
	if o.Pkg {
		format += "%s"
		vals = append(vals, s.Pkg)
	}

	if o.Distro {
		format += "%s"
		vals = append(vals, s.Distro)
	}

	if o.Arch {
		format += "%s"
		vals = append(vals, s.Arch)
	}

	if o.Repo {
		format += "%s"
		vals = append(vals, s.Repo)
	}

	if o.Commit {
		format += "%s"
		vals = append(vals, s.Commit)
	}

	if o.Tag {
		format += "%s"
		vals = append(vals, s.Tag)
	}

	if o.Revision {
		format += "%s"
		vals = append(vals, s.Revision)
	}

	h := strings.ReplaceAll(fmt.Sprintf(format, vals...), "/", "")
	return h, nil
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
