package archive

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/Azure/moby-packaging/targets"
)

var (
	alphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

type Spec struct {
	Pkg      string `json:"package"`
	Distro   string `json:"distro"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo" hash:"ignore"`
	Commit   string `json:"commit" hash:"ignore"`
	Tag      string `json:"tag"`
	Revision string `json:"revision"`
}

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
func (s *Spec) Hash() (string, error) {
	v := reflect.ValueOf(s)
	w := v.Elem()

	ret := make([]string, 0, w.NumField())
	for i := 0; i < w.NumField(); i++ {
		fieldTag := w.Type().Field(i).Tag

		str := w.Field(i).String()
		if fieldTag.Get("hash") == "ignore" || str == "" {
			continue
		}

		str = alphanumeric.ReplaceAllString(str, "_")
		ret = append(ret, str)
	}

	retStr := strings.Join(ret, ".")

	return retStr, nil
}

func (s *Spec) Dir(rootDir string) string {
	pkgOS := s.OS()
	sanitizedArch := strings.ReplaceAll(s.Arch, "/", "_")
	osArchDir := fmt.Sprintf("%s_%s", pkgOS, sanitizedArch)
	artifactDir := filepath.Join(rootDir, s.Distro, osArchDir)

	return artifactDir
}

func (s *Spec) Basename() (string, error) {
	n, ok := targets.StaticTargetAttributes[s.Distro]
	if !ok {
		return "", fmt.Errorf("Distro not understood: '%s'", s.Distro)
	}

	o := n.OsComponent

	extension := n.Extension
	version := n.VersionComponent
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
