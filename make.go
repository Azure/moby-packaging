package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/targets"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

func main() {
	outDir := flag.String("output", "bundles", "Output directory for built packages (note the distro name will be appended to this path)")
	buildSpec := flag.String("build-spec", "", "Location of the build spec json file")

	flag.Parse()

	var specs []*archive.Spec

	if spec, err := readBuildSpec(*buildSpec); err != nil {
		var err2 error
		if specs, err2 = readBuildSpecMulti(*buildSpec); err2 != nil {
			fmt.Fprintln(os.Stderr, "Could not parse build spec as either single or multi-spec:", errors.Join(err, err2))
			os.Exit(1)
		}
	} else {
		specs = append(specs, spec)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer cancel()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer client.Close()

	go func() {
		<-ctx.Done()
		client.Close()
	}()

	grp, ctx := errgroup.WithContext(ctx)
	for _, spec := range specs {
		spec := spec
		grp.Go(func() error {
			out, err := do(ctx, client, spec)
			if err != nil {
				return err
			}
			_, err = out.Export(ctx, *outDir)
			return err
		})
	}

	if err := grp.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}

func readBuildSpecMulti(filename string) ([]*archive.Spec, error) {
	if filename == "" {
		return nil, fmt.Errorf("no build spec file specified")
	}

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec []*archive.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func readBuildSpec(filename string) (*archive.Spec, error) {
	if filename == "" {
		return nil, fmt.Errorf("no build spec file specified")
	}

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec archive.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func do(ctx context.Context, client *dagger.Client, cfg *archive.Spec) (*dagger.Directory, error) {
	if cfg.Arch == "" {
		p, err := client.DefaultPlatform(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not determine default platform: %w", err)
		}

		_, a, ok := strings.Cut(string(p), "/")
		if !ok {
			return nil, fmt.Errorf("unexpected platform format: %q", p)
		}
		cfg.Arch = a
	}

	targetOs := "linux"
	if cfg.Distro == "windows" {
		targetOs = "windows"
	}
	platform := dagger.Platform(fmt.Sprintf("%s/%s", targetOs, cfg.Arch))

	getGoVersion, ok := targets.GetGoVersionForPackage[cfg.Pkg]
	if !ok {
		return nil, fmt.Errorf("unknown package: %q", cfg.Pkg)
	}
	goVersion := getGoVersion(cfg)

	target, err := targets.GetTarget(ctx, cfg.Distro, client, platform, goVersion)
	if err != nil {
		return nil, err
	}

	return target.Make(cfg, packageDir(client, cfg.Pkg), hackCrossDir(client))
}
