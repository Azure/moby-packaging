package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/moby-packaging/pkg/archive"
)

func main() {
	s := archive.Spec{}
	flag.StringVar(&s.Pkg, "project", "", "name of the project")
	flag.StringVar(&s.Distro, "distro", "", "distro of artifact")
	flag.StringVar(&s.Arch, "arch", "", "arch of artifact")
	flag.StringVar(&s.Tag, "tag", "", "tag of artifact")
	flag.StringVar(&s.Revision, "revision", "", "revision")
	flag.Parse()

	h, err := s.Hash()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("%s", h)
}
