package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/pborman/getopt/v2"
)

type args struct {
	bundleDir string
	specFile  string
	cmd       string
}

var (
	a = args{}
)

func init() {
	getopt.FlagLong(&a.bundleDir, "bundle-dir", 'b', "base directory of bundled files")
	getopt.FlagLong(&a.specFile, "spec-file", 's', "spec file to calculate path")
}
func main() {

	fmt.Println("args:", getopt.Args())

	if err := do("dir", a); err != nil {
		panic(err)
	}
}

func do(cmd string, a args) error {
	fmt.Println("cmd:", cmd)
	fmt.Println("spec file:", a.specFile)
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
	case "", "dir":
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
	}

	fmt.Printf("%s", p)

	return nil
}
