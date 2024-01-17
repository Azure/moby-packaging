package targets

import (
	"context"
	"fmt"
	"testing"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/goversion"
	"golang.org/x/sync/errgroup"
)

type testLogWriter struct {
	t *testing.T
}

func (t *testLogWriter) Write(b []byte) (int, error) {
	t.t.Helper()
	t.t.Log(string(b))
	return len(b), nil
}

func TestApt(t *testing.T) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(&testLogWriter{t}))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Jammy(ctx, client, platform, goversion.DefaultVersion)
	if err != nil {
		t.Fatal(err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	for _, pkg := range BaseDebPackages {
		pkg := pkg
		eg.Go(func() error {
			c := c.WithExec([]string{"/usr/bin/dpkg", "-s", pkg})
			out, err := c.Container().Stderr(ctx)
			if err != nil {
				return fmt.Errorf("error checking if package is installed: %s: %v:\n%s", pkg, err, out)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
