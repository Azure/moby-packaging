package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/moby-packaging/pkg/archive"
)

func hash(s *archive.Spec) (string, error) {
	h := strings.ReplaceAll(fmt.Sprintf("%s%s%s%s%s", s.Pkg, s.Distro, s.Arch, s.Tag, s.Revision), "/", "")
	return h, nil
}

func main() {
	s := archive.Spec{}
	flag.StringVar(&s.Pkg, "project", "", "name of the project")
	flag.StringVar(&s.Distro, "distro", "", "distro of artifact")
	flag.StringVar(&s.Arch, "arch", "", "arch of artifact")
	flag.StringVar(&s.Tag, "tag", "", "tag of artifact")
	flag.StringVar(&s.Revision, "revision", "", "revision")
	flag.Parse()

	h, err := hash(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("%s", h)
}
