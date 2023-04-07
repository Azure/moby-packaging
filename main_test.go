package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"packaging/pkg/build"
	"testing"

	"dagger.io/dagger"
)

var (
	client    *dagger.Client
	buildSpec *build.Spec
)

func TestMain(m *testing.M) {
	specFile := flag.String("build-spec", "", "distro to test")

	flag.Parse()
	var err error
	buildSpec, err = readBuildSpec(*specFile)
	if err != nil {
		panic(err)
	}
	fmt.Println(buildSpec)

	client, err = dagger.Connect(context.Background(), dagger.WithWorkdir("../"))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	client.Close()
	os.Exit(code)
}
