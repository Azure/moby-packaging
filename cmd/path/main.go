package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Azure/moby-packaging/pkg/archive"
)

type args struct {
	bundleDir string
	specFile  string
}

func main() {
	a := args{}

	if len(os.Args) < 2 {
		panic("first arg must be 'dir', 'full-path', or 'basename'")
	}

	fullPath := flag.NewFlagSet("path", flag.ExitOnError)
	fullPath.StringVar(&a.bundleDir, "bundle-dir", "", "base directory of bundled files")
	fullPath.StringVar(&a.specFile, "spec-file", "", "path of spec file")
	fullPath.Parse(os.Args[2:])

	if err := do(os.Args[1], a); err != nil {
		panic(err)
	}
}

func do(cmd string, a args) error {
	b, err := os.ReadFile(a.specFile)
	if err != nil {
		return err
	}

	var s archive.Spec
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	var p string
	switch cmd {
	case "dir":
		p = s.Dir(a.bundleDir)
	case "full-path":
		var err error
		p, err = s.FullPath(a.bundleDir)
		if err != nil {
			return err
		}
	case "basename":
		var err error
		p, err = s.Basename()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("command not recognized")
	}

	fmt.Printf("%s", p)

	return nil
}
