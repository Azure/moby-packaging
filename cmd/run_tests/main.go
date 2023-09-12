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
	InstructionsPath string
	BundleDirPath    string
}

func main() {
	args := Args{}
	flag.StringVar(&args.InstructionsPath, "instructions-file", "", "path of the pipeline instructions file to be used")
	flag.StringVar(&args.BundleDirPath, "bundle-dir", "", "path of the bundle dir to test")
	flag.Parse()

	if err := runTest(args); err != nil {
		panic(err)
	}

}

func runTest(args Args) error {
	b, err := os.ReadFile(args.InstructionsPath)
	if err != nil {
		return err
	}

	var pi archive.PipelineInstructions
	if err := json.Unmarshal(b, &pi); err != nil {
		return err
	}

	transformed := strings.TrimPrefix(pi.Pkg, "moby-")
	transformed = strings.ToUpper(transformed)
	transformed = strings.ReplaceAll(transformed, "-", "_")

	pkgOs, ok := osMap[pi.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	pv, ok := versionMap[pi.Distro]
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
		/* 1 */ pi.Distro,
		/* 2 */ pi.Arch,
		/* 3 */ transformed,
		/* 4 */ pi.Commit,
		/* 5 */ pi.Tag,
		/* 6 */ pi.Revision,
		/* 7 */ pv,
	)

	tagRevision := fmt.Sprintf("%s-%s", pi.Tag, pi.Revision)
	pkgVer := fmt.Sprintf("%s.%s", tagRevision, pv)

	if pkgOs == debian || pkgOs == ubuntu {
		pkgVer = fmt.Sprintf("%[1]s-%[2]s%[3]su%[4]s",
			/* 1 */ pi.Tag,
			/* 2 */ pkgOs,
			/* 3 */ pv,
			/* 4 */ pi.Revision,
		)
	}

	cmd := exec.Command("/usr/bin/make", "test", fmt.Sprintf("OUTPUT=%s", args.BundleDirPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DISTRO=%s", pi.Distro))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TARGETARCH=%s", pi.Arch))
	cmd.Env = append(cmd.Env, "INCLUDE_TESTING=0")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_COMMIT=%s", transformed, pi.Commit))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_VERSION=%s", transformed, tagRevision))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_%s_PACKAGE_VERSION=%s", transformed, pkgVer))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	h, err := pi.Hash(archive.HashOptions{
		Pkg:    true,
		Distro: true,
		Arch:   true,
	})
	if err != nil {
		return err
	}

	return nil
}
