package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/targets"
	"golang.org/x/sys/unix"
)

func main() {
	outDir := flag.String("output", "bundles", "Output directory for built packages (note the distro name will be appended to this path)")
	buildSpec := flag.String("build-spec", "", "Location of the build spec json file")

	flag.Parse()

	spec, err := readBuildSpec(*buildSpec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read or parse build spec file")
		os.Exit(1)
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

	out, err := do(ctx, client, spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	if _, err := out.Export(ctx, spec.Dir(*outDir)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
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
	out := target.Make(cfg, packageDir(client, cfg.Pkg), hackCrossDir(client))
	return out, nil
}
