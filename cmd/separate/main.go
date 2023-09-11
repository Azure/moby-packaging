package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var (
	r = regexp.MustCompile(`.*\.(zip|rpm|deb)`)
)

type Mapping struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

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
	m := []Mapping{}
	dstDir, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}

	if err := filepath.WalkDir(args[0], func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if !r.MatchString(path) {
			return nil
		}

		// Move the file, recording source and destination
		src, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		base := filepath.Base(src)
		dst := filepath.Join(dstDir, base)

		mapping := Mapping{
			Src: src,
			Dst: dst,
		}

		m = append(m, mapping)

		return nil
	}); err != nil {
		return err
	}
	return nil
}
