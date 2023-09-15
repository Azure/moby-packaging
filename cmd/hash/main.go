package main

import (
	"flag"
	"fmt"

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

	h := s.NameTagRevision()
	fmt.Printf("%s", h)
}
