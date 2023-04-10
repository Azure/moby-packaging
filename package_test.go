package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"dagger.io/dagger"
	"github.com/cpuguy83/go-docker"
	"github.com/cpuguy83/go-docker/container"
	"github.com/cpuguy83/go-docker/container/containerapi"
	"github.com/cpuguy83/go-docker/container/containerapi/mount"
	"github.com/cpuguy83/go-docker/image"
	"golang.org/x/sync/errgroup"
)

var (
	GoVersion = "1.19.5"
	GoRef     = path.Join("mcr.microsoft.com/oss/go/microsoft/golang:" + GoVersion)

	//go:embed tests/entrypoint.sh
	systemdScript string
)

func TestPackage(t *testing.T) {
	t.Run(fmt.Sprintf("test %s on %s", buildSpec.Pkg, buildSpec.Distro), func(t *testing.T) {
		ctx := context.Background()
		// set up the daemon container
		getContainer, ok := distros[buildSpec.Distro]
		if !ok {
			t.Fatalf("unknown distro: %s", buildSpec.Distro)
		}

		client := getClient(t)

		c := getContainer(client).
			WithNewFile("/entrypoint.sh", dagger.ContainerWithNewFileOpts{Contents: systemdScript, Permissions: 0o755})

		dir := t.TempDir()
		imgPath := filepath.Join(dir, "img.tar")

		_, err := c.Export(ctx, imgPath)
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(imgPath)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		docker := docker.NewClient()
		var loadRef string
		err = docker.ImageService().Load(ctx, f, func(config *image.LoadConfig) error {
			config.ConsumeProgress = func(ctx context.Context, rdr io.Reader) error {
				type progress struct {
					Stream string
				}
				var p progress

				for {
					if err := json.NewDecoder(rdr).Decode(&p); err != nil {
						if err == io.EOF {
							break
						}
						return err
					}
					if p.Stream == "" {
						continue
					}

					_, ref, ok := strings.Cut(p.Stream, ":")
					if !ok {
						continue
					}
					loadRef = strings.TrimSpace(ref)
					break

				}
				io.Copy(io.Discard, rdr)
				if loadRef == "" {
					return fmt.Errorf("failed to find load ref")
				}
				return nil
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		withHC := container.WithCreateHostConfig(
			containerapi.HostConfig{
				Privileged: true,
				AutoRemove: true,
				Mounts: []mount.Mount{
					{
						Type:     mount.TypeBind,
						Source:   "/sys/fs/cgroup",
						Target:   "/sys/fs/cgroup",
						ReadOnly: true,
					},
				},
			},
		)

		ctr, err := docker.ContainerService().Create(ctx, loadRef, withHC, container.WithCreateConfig(
			containerapi.Config{
				Image:      loadRef,
				Entrypoint: []string{"/entrypoint.sh"},
				StopSignal: "SIGRTMIN+3",
				Env:        []string{"container=docker"},
			},
		))
		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			docker.ContainerService().Remove(ctx, ctr.ID(), container.WithRemoveForce)
		})

		ws, err := ctr.Wait(ctx, container.WithWaitCondition(container.WaitConditionNextExit))
		if err != nil {
			t.Fatal(err)
		}

		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			code, err := ws.ExitCode()
			if err != nil {
				return err
			}
			if code != 0 {
				ctr.Logs(ctx, func(cfg *container.LogReadConfig) {
					cfg.Stdout = &testWriter{t: t, prefix: "logs-stdout"}
					cfg.Stderr = &testWriter{t: t, prefix: "logs-stderr"}
				})
				return fmt.Errorf("exit code: %d", code)
			}
			return nil
		})
		defer func() {
			ctr.Stop(ctx)
			if err := eg.Wait(); err != nil {

				t.Error(err)
			}
		}()

		if err := ctr.Start(ctx); err != nil {
			t.Fatal(err)
		}

		ep, err := ctr.Exec(ctx, func(config *container.ExecConfig) {
			config.Cmd = []string{"/opt/moby/install.sh"}
			config.Privileged = true
			config.Stdout = &testWriter{t: t, prefix: "exec-stdout"}
			config.Stderr = &testWriter{t: t, prefix: "exec-stderr"}
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := ep.Start(ctx); err != nil {
			t.Fatal(err)
		}
	})
}
