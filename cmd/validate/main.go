package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Azure/moby-packaging/pkg/archive"
)

func main() {
	pkgName := ""
	if len(os.Args) > 1 {
		pkgName = os.Args[1]
	}
	r := bufio.NewReader(os.Stdin)

	j, err := r.ReadBytes(byte(0))
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := validate(j, pkgName); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func validate(j []byte, pkgName string) error {
	specs := []archive.Spec{}

	if err := json.Unmarshal(j, &specs); err != nil {
		return err
	}

	for i := range specs {
		if pn := specs[i].Pkg; pn != pkgName {
			return fmt.Errorf("package name does not match: '%s' vs '%s'", pkgName, pn)
		}

		s := []string{
			specs[i].Arch,
			specs[i].Commit,
			specs[i].Distro,
			specs[i].Pkg,
			specs[i].Repo,
			specs[i].Revision,
			specs[i].Tag,
		}

		for _, ss := range s {
			if ss == "" {
				return fmt.Errorf("blank value")
			}
		}
	}

	return nil
}
