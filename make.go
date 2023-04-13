package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"packaging/targets"
	"path/filepath"
	"strings"
	"time"

	"packaging/pkg/archive"
	"packaging/pkg/build"

	"dagger.io/dagger"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
)

func main() {
	flags := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)

	go func() {
		server := &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 3 * time.Second,
		}
		err := server.ListenAndServe()
		fmt.Fprintln(os.Stderr, err)
	}()

	buildSpec := flags.String("build-spec", "", "Location of the build spec json file")
	pkgDef := flags.String("package-definition", "", "Location of the package definition yaml file")

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	b, err := os.ReadFile(*pkgDef)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad filename")
		flag.Usage()
		os.Exit(1)
	}

	var p archive.NewArchive
	if err := yaml.Unmarshal(b, &p); err != nil {
		fmt.Fprintln(os.Stderr, "can't read yaml")
		flag.Usage()
		os.Exit(1)
	}

	targets.Definition = p

	if *buildSpec == "" {
		fmt.Fprintln(os.Stderr, "no build spec provided")
		flag.Usage()
		os.Exit(1)
	}

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
		os.Exit(1)
	}
	defer client.Close()

	if spec.Arch == "" {
		p, err := client.DefaultPlatform(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not get default platform:", err)
			os.Exit(1)
		}
		_, a, ok := strings.Cut(string(p), "/")
		if !ok {
			fmt.Fprintln(os.Stderr, "got unexpected platform:", p)
			os.Exit(2)
		}
		spec.Arch = a
	}

	if err := do(ctx, client, spec); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readBuildSpec(filename string) (*build.Spec, error) {
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

func do(ctx context.Context, client *dagger.Client, cfg *build.Spec) error {
	platform := dagger.Platform(fmt.Sprintf("%s/%s", cfg.OS, cfg.Arch))

	target, err := targets.GetTarget(cfg.Distro)(ctx, client, platform)
	if err != nil {
		return err
	}
	out := target.Make(cfg)

	_, err = out.Export(ctx, filepath.Join("bundles", cfg.Distro))
	return err
}
