package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	gofs "io/fs"
	"path"
	"path/filepath"
	"testing"

	"dagger.io/dagger"
)

var (
	GoVersion = "1.19.5"
	GoRef     = path.Join("mcr.microsoft.com/oss/go/microsoft/golang:" + GoVersion)

	//go:embed tests/entrypoint.sh
	systemdScript string

	//go:embed cmd/testingapi/testingapi.service
	systemdUnit string

	//go:embed go.mod go.sum all:cmd
	daemonDir embed.FS
)

func TestPackage(t *testing.T) {
	t.Run(fmt.Sprintf("test %s on %s", buildSpec.Pkg, buildSpec.Distro), func(t *testing.T) {
		ctx := context.Background()
		// set up the daemon container
		getContainer, ok := distros[buildSpec.Distro]
		if !ok {
			t.Fatalf("unknown distro: %s", buildSpec.Distro)
		}

		build := getContainer(client)
		var err error
		build, err = installGo(ctx, build)
		if err != nil {
			t.Fatal(err)
		}

		daemonBuildDir := client.Directory()
		daemonBuildDir, err = GoFSToDagger(daemonDir, daemonBuildDir, func(s string) string {
			return filepath.Join("build", s)
		})
		if err != nil {
			t.Fatal(err)
		}

		daemonBin := build.
			WithDirectory("/build", daemonBuildDir).
			WithWorkdir("/build/build").
			WithExec([]string{"go", "build", "-o", "/tmp/testingapi", "./cmd/testingapi"}).
			File("/tmp/testingapi")

		s, err := getContainer(client).
			WithFile("/usr/local/bin/testingapi", daemonBin).
			WithMountedTemp("/tmp").
			WithMountedTemp("/run").
			WithMountedTemp("/run/lock").
			// WithMountedDirectory("/sys/fs/cgroup", cg).
			WithNewFile("/etc/systemd/system/testingapi.service", dagger.ContainerWithNewFileOpts{
				Contents:    systemdUnit,
				Permissions: 0o644,
			}).
			WithNewFile("/entrypoint.sh", dagger.ContainerWithNewFileOpts{Contents: systemdScript, Permissions: 0o755}).
			WithExec([]string{"/entrypoint.sh"}, dagger.ContainerWithExecOpts{InsecureRootCapabilities: true}).Stdout(ctx)

		fmt.Println(s)
		if err != nil {
			t.Error(err)
		}

	})
}

// GoFSToDa// RewriteFn is a function that can be used to rewrite the path of a file or directory
// It is used by GoFSToDagger to allow callers to pass in a function that can rewrite the path.
type RewriteFn func(string) string

func GoFSToDagger(fs gofs.FS, dir *dagger.Directory, rewrite RewriteFn) (*dagger.Directory, error) {
	err := gofs.WalkDir(fs, ".", func(path string, d gofs.DirEntry, err error) error {
		// fmt.Println(path)
		if err != nil {
			return err
		}

		rewritten := rewrite(path)
		if rewritten == "" {
			return nil
		}

		if path == "." {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		perm := fi.Mode().Perm()
		if d.IsDir() {
			dir = dir.WithNewDirectory(rewritten, dagger.DirectoryWithNewDirectoryOpts{Permissions: int(perm)})
			return nil
		}

		b, err := gofs.ReadFile(fs, path)
		if err != nil {
			return err
		}

		dir = dir.WithNewFile(rewritten, string(b), dagger.DirectoryWithNewFileOpts{Permissions: int(perm)})
		return nil
	})
	return dir, err
}

func installGo(ctx context.Context, c *dagger.Container) (*dagger.Container, error) {
	dir := c.From(GoRef).Directory("/usr/local/go")
	pathEnv, err := c.EnvVariable(ctx, "PATH")
	if err != nil {
		return nil, fmt.Errorf("error getting PATH: %w", err)
	}
	if pathEnv == "" {
		return nil, fmt.Errorf("PATH is empty")
	}

	return c.WithDirectory("/usr/local/go", dir).
		WithEnvVariable("PATH", "/go/bin:/usr/local/go/bin:"+pathEnv).
		WithEnvVariable("GOROOT", "/usr/local/go").
		WithEnvVariable("GOPATH", "/go"), nil
}
