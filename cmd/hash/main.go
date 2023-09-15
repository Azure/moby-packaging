package main

import (
	"flag"
	"fmt"

	"github.com/Azure/moby-packaging/pkg/archive"
)

func main() {
	s := archive.Spec{}
	flag.StringVar(&s.Pkg, "project", "", "name of the project")
	flag.StringVar(&s.Revision, "revision", "", "revision")
	flag.StringVar(&s.Tag, "tag", "", "tag of artifact")
	flag.Parse()

	h := s.NameTagRevision()
	fmt.Printf("%s", h)
}
