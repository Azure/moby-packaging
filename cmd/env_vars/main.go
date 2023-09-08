package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/Azure/moby-packaging/pkg/archive"
)

const (
	debian  = "debian"
	ubuntu  = "ubuntu"
	makebin = "make"
)

var (
	osMap = map[string]string{
		"bookworm": "debian",
		"bullseye": "debian",
		"buster":   "debian",
		"focal":    "ubuntu",
		"jammy":    "ubuntu",
		"rhel9":    "el9",
		"rhel8":    "el8",
		"centos7":  "el7",
		"mariner2": "cm2",
	}

	versionMap = map[string]string{
		"bookworm": "12",
		"bullseye": "11",
		"buster":   "10",
		"focal":    "20.04",
		"jammy":    "22.04",
		"rhel9":    "el9",
		"rhel8":    "el8",
		"centos7":  "el7",
		"mariner2": "cm2",
	}
)

func main() {
	if len(os.Args) < 3 {
		panic("the first arg must be the path of the spec file, the second arg must be the bundle dir")
	}

	if err := do(); err != nil {
		panic(err)
	}
}

func do() error {
	specPath := os.Args[1]

	b, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	var s archive.Spec
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	transformed := strings.TrimPrefix(s.Pkg, "moby-")
	transformed = strings.ToUpper(transformed)
	transformed = strings.ReplaceAll(transformed, "-", "_")

	pkgOs, ok := osMap[s.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	pkgVer, ok := versionMap[s.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	fmt.Printf(`
export DISTRO=%[1]s
export TARGETARCH=%[2]s
export INCLUDE_TESTING=[0]
export TEST_%[3]s_COMMIT=%[4]s
export TEST_%[3]s_VERSION=%[5]s-%[6]s
export TEST_%[3]s_PACKAGE_VERSION=%[5]s-%[6]s.%[7]s
`,
		/* 1 */ s.Distro,
		/* 2 */ s.Arch,
		/* 3 */ transformed,
		/* 4 */ s.Commit,
		/* 5 */ s.Tag,
		/* 6 */ s.Revision,
		/* 7 */ pkgVer,
	)

	tagRevision := fmt.Sprintf("%s-%s", s.Tag, s.Revision)

	os.Setenv("DISTRO", s.Distro)
	os.Setenv("TARGETARCH", s.Arch)
	os.Setenv("INCLUDE_TESTING", "0")
	os.Setenv(fmt.Sprintf("TEST_%s_COMMIT", transformed), s.Commit)
	os.Setenv(fmt.Sprintf("TEST_%s_VERSION", transformed), tagRevision)
	os.Setenv(fmt.Sprintf("TEST_%s_PACKAGE_VERSION", transformed), fmt.Sprintf("%s.%s", tagRevision, pkgVer))

	if pkgOs == debian || pkgOs == ubuntu {
		os.Setenv(
			fmt.Sprintf("TEST_%s_PACKAGE_VERSION", transformed),
			fmt.Sprintf("%[1]s-%[2]s%[3]su%[4]s",
				/* 1 */ s.Tag,
				/* 2 */ pkgOs,
				/* 3 */ pkgVer,
				/* 4 */ s.Revision,
			),
		)
	}

	bin, err := exec.LookPath(makebin)
	if err != nil {
		return err
	}

	env := os.Environ()
	return syscall.Exec(bin, []string{makebin, "test", fmt.Sprintf("OUTPUT=%s", os.Args[2])}, env)
}
