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

	pkgOs, ok := archive.OSMap[s.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	pv, ok := archive.VersionMap[s.Distro]
	if !ok {
		return fmt.Errorf("unrecognized distro: %s", pkgOs)
	}

	fmt.Fprintf(os.Stderr, "%+v\n%s", s, pv)

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

	runMake, err := exec.LookPath(makebin)
	if err != nil {
		return err
	}

	cmd := exec.Command(runMake, "test", fmt.Sprintf("OUTPUT=%s", args.BundleDirPath))
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

	return nil
}
