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
	r := bufio.NewReader(os.Stdin)

	j, err := r.ReadBytes(byte(0))
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := validate(j); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	return
}

func validate(j []byte) error {
	specs := []archive.Spec{}

	if err := json.Unmarshal(j, &specs); err != nil {
		return err
	}

	for i := range specs {
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
