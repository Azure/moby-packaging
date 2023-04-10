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
	buildSpec *build.Spec
)

func getClient(t *testing.T) *dagger.Client {
	t.Helper()
	client, err := dagger.Connect(context.Background(), dagger.WithWorkdir(os.TempDir()), dagger.WithLogOutput(&testWriter{t: t, prefix: "dagger"}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Log(err)
		}
	})
	return client
}

type testWriter struct {
	t      *testing.T
	prefix string
}

func (d *testWriter) Write(p []byte) (n int, err error) {
	d.t.Log(d.prefix + ": " + string(p))
	return len(p), nil
}

func (d *testWriter) Close() error {
	return nil
}

func TestMain(m *testing.M) {
	specFile := flag.String("build-spec", "", "distro to test")
	flag.Parse()

	var err error
	buildSpec, err = readBuildSpec(*specFile)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, *buildSpec)

	os.Exit(m.Run())
}
