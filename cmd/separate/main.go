package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Azure/moby-packaging/pkg/archive"
)

var (
	r = regexp.MustCompile(`.*\.(zip|rpm|deb)`)
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "first arg must be file containing a list of specs, second arg is output directory")
		os.Exit(1)
	}

	if err := do(args); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

func do(args []string) error {
    specs := []archive.Spec{}
    b, err := os.ReadFile(args[0])
    if err != nil {
        return err
    }

    if err := json.Unmarshal(b, &specs); err != nil {
        return err
    }

    for _, spec := range specs {
        path := 
    }

    return nil
}
