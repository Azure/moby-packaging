package archive

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

	ExtensionMap = map[string]string{
		"bionic":   "deb",
		"bookworm": "deb",
		"bullseye": "deb",
		"buster":   "deb",
		"focal":    "deb",
		"jammy":    "deb",
		"noble":    "deb",
		"rhel9":    "rpm",
		"rhel8":    "rpm",
		"centos7":  "rpm",
		"mariner2": "rpm",
		"windows":  "zip",
	}

	OSMap = map[string]string{
		"bookworm": "debian",
		"bullseye": "debian",
		"buster":   "debian",
		"bionic":   "ubuntu",
		"focal":    "ubuntu",
		"jammy":    "ubuntu",
		"noble":    "ubuntu",
		"rhel9":    "el9",
		"rhel8":    "el8",
		"centos7":  "el7",
		"mariner2": "cm2",
		"windows":  "windows",
	}

	VersionMap = map[string]string{
		"bookworm": "12",
		"bullseye": "11",
		"buster":   "10",
		"bionic":   "18.04",
		"focal":    "20.04",
		"jammy":    "22.04",
		"noble":    "24.04",
		"rhel9":    "el9",
		"rhel8":    "el8",
		"centos7":  "el7",
		"mariner2": "cm2",
	}
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

// This function calculates the storage path for a package in the prod storage
// container.
func (spec *Spec) StoragePath() (string, error) {
	pkg := spec.Pkg
	pkgOS := spec.OS()
	version := fmt.Sprintf("%s+azure", spec.Tag)
	distro := spec.Distro
	sanitizedArch := strings.ReplaceAll(spec.Arch, "/", "_")

	base, err := spec.Basename()
	if err != nil {
		return "", err
	}

	storagePath := fmt.Sprintf("%s/%s/%s/%s_%s/%s", pkg, version, distro, pkgOS, sanitizedArch, base)

	return storagePath, nil
}

// This logic is arbitrary, but the output must be reproducible. This is used
// to generate filenames for artifacts.
func (s *Spec) NameTagRevision() string {
	pkg := s.Pkg
	tag := s.Tag
	rev := s.Revision

	for _, ptr := range []*string{&pkg, &tag, &rev} {
		*ptr = nonAlnum.ReplaceAllString(*ptr, "_")
	}

	return fmt.Sprintf("%s.%s.%s", pkg, tag, rev)
}

// Our pipelines have historically used opinionated filesystem layouts to place
// artifacts in a consistent location. This method will determine the directory
// structure for all path components except the basename of the artifact
// produced by this build spec definition. Because the base directory can be
// different depending on the situation, it is supplied as an argument. Use "."
// as the rootDir in order to use a relative path.
func (s *Spec) Dir(rootDir string) string {
	pkgOS := s.OS()
	sanitizedArch := strings.ReplaceAll(s.Arch, "/", "_")
	osArchDir := fmt.Sprintf("%s_%s", pkgOS, sanitizedArch)
	artifactDir := filepath.Join(rootDir, s.Distro, osArchDir)

	return artifactDir
}

// There are semantic rules on the naming of packages for both debian- and rpm-
// based repositories. This method will generate the basename of the package
// name, according to those semantic rules, based on the information in the
// build spec.
func (s *Spec) Basename() (string, error) {
	o, ok := OSMap[s.Distro]
	if !ok {
		return "", fmt.Errorf("Distro not understood: '%s'", s.Distro)
	}

	extension := ExtensionMap[s.Distro]
	version := VersionMap[s.Distro]
	sanitizedArch := strings.ReplaceAll(s.Arch, "/", "")
	str := ""

	switch o {
	case "debian", "ubuntu":
		str = fmt.Sprintf("%[1]s_%[2]s-%[3]s%[4]su%[5]s_%[6]s.%[7]s",
			/* 1 */ s.Pkg,
			/* 2 */ s.Tag,
			/* 3 */ o,
			/* 4 */ version,
			/* 5 */ s.Revision,
			/* 6 */ sanitizedArch,
			/* 7 */ extension,
		)
	case "windows":
		str = fmt.Sprintf("%[1]s-%[2]s+azure-u%[3]s.%[4]s.%[5]s",
			/* 1 */ s.Pkg,
			/* 2 */ s.Tag,
			/* 3 */ s.Revision,
			/* 4 */ sanitizedArch,
			/* 5 */ extension,
		)
	default:
		arch, ok := rpmArchMap[s.Arch]
		if !ok {
			arch = s.Arch
		}
		str = fmt.Sprintf("%[1]s-%[2]s-%[3]s.%[4]s.%[5]s.%[6]s",
			/* 1 */ s.Pkg,
			/* 2 */ s.Tag,
			/* 3 */ s.Revision,
			/* 4 */ o,
			/* 5 */ arch,
			/* 6 */ extension,
		)
	}

	return str, nil
}

// This method is provided for convenience, simply combinging `.Dir()` and
// `.Basename()`. See the documentation for those methods for more information.
func (s *Spec) FullPath(rootDir string) (string, error) {
	f, err := s.Basename()
	if err != nil {
		return "", err
	}
	return filepath.Join(s.Dir(rootDir), f), nil
}

func (s *Spec) OS() string {
	if s.Distro == "windows" {
		return "windows"
	}

	return "linux"
}
