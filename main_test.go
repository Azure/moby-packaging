package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"testing"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
)

var (
	buildSpec *archive.Spec
	flDebug   bool
	signalCtx context.Context
)

func getClient(ctx context.Context, t *testing.T) *dagger.Client {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	client, err := dagger.Connect(ctx, dagger.WithWorkdir(wd), dagger.WithLogOutput(&testWriter{t: t, prefix: "dagger"}))
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
	flag.BoolVar(&flDebug, "debug", false, "enable debug logging")
	flag.Parse()

	var cancel func()
	signalCtx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if *specFile != "" {
		var err error
		buildSpec, err = readBuildSpec(*specFile)
		if err != nil {
			panic(err)
		}
	}

	os.Exit(m.Run())
}
