package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"packaging/targets"
	"path/filepath"
	"strings"

	"packaging/pkg/build"

	"dagger.io/dagger"
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

	out, err := do(ctx, client, spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	if _, err := out.Export(ctx, filepath.Join(*outDir, spec.Distro)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
}

func readBuildSpec(filename string) (*build.Spec, error) {
	if filename == "" {
		return nil, fmt.Errorf("no build spec file specified")
	}

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec build.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func do(ctx context.Context, client *dagger.Client, cfg *build.Spec) (*dagger.Directory, error) {
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

	platform := dagger.Platform(fmt.Sprintf("%s/%s", cfg.OS, cfg.Arch))

	target, err := targets.GetTarget(cfg.Distro)(ctx, client, platform)
	if err != nil {
		return nil, err
	}
	out := target.Make(cfg)
	return out, nil
}
