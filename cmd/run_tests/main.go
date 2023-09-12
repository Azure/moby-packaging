package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

type Args struct {
	SpecPath      string
	BundleDirPath string
}

func main() {
	args := Args{}
	flag.StringVar(&args.SpecPath, "spec-file", "", "path of the pipeline instructions file to be used")
	flag.StringVar(&args.BundleDirPath, "bundle-dir", "", "path of the bundle dir to test")
	flag.Parse()

	if err := runTest(args); err != nil {
		panic(err)
	}

}

func runTest(args Args) error {
	b, err := os.ReadFile(args.SpecPath)
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

	pv, ok := versionMap[s.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	fmt.Printf(`
DISTRO=%[1]s
TARGETARCH=%[2]s
INCLUDE_TESTING=[0]
TEST_%[3]s_COMMIT=%[4]s
TEST_%[3]s_VERSION=%[5]s-%[6]s
TEST_%[3]s_PACKAGE_VERSION=%[5]s-%[6]s.%[7]s
`,
		/* 1 */ s.Distro,
		/* 2 */ s.Arch,
		/* 3 */ transformed,
		/* 4 */ s.Commit,
		/* 5 */ s.Tag,
		/* 6 */ s.Revision,
		/* 7 */ pv,
	)

	tagRevision := fmt.Sprintf("%s-%s", s.Tag, s.Revision)
	pkgVer := fmt.Sprintf("%s.%s", tagRevision, pv)

	if pkgOs == debian || pkgOs == ubuntu {
		pkgVer = fmt.Sprintf("%[1]s-%[2]s%[3]su%[4]s",
			/* 1 */ s.Tag,
			/* 2 */ pkgOs,
			/* 3 */ pv,
			/* 4 */ s.Revision,
		)
	}

	cmd := exec.Command("/usr/bin/make", "test", fmt.Sprintf("OUTPUT=%s", args.BundleDirPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DISTRO=%s", s.Distro))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TARGETARCH=%s", s.Arch))
	cmd.Env = append(cmd.Env, "INCLUDE_TESTING=0")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_COMMIT=%s", transformed, s.Commit))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_VERSION=%s", transformed, tagRevision))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_PACKAGE_VERSION=%s", transformed, pkgVer))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// h, err := pi.Hash()
	// if err != nil {
	// 	return err
	// }

	return nil
}
