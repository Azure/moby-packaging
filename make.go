package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
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

	targetOS := "linux"
	if spec.Distro == "windows" {
		targetOS = "windows"
	}

	sanitizedArch := strings.ReplaceAll(spec.Arch, "/", "_")
	subDir := fmt.Sprintf("%s_%s", targetOS, sanitizedArch)

	artifactDir := filepath.Join(*outDir, spec.Distro, subDir)
	if _, err := out.Export(ctx, artifactDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	artifact, err := findArtifact(artifactDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}

	absPath, err := filepath.Abs(filepath.Dir(artifact))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(6)
	}

	pi := archive.PipelineInstructions{
		Spec:                *spec,
		SpecPath:            *buildSpec,
		Basename:            filepath.Base(artifact),
		OriginalArtifactDir: filepath.Dir(absPath),
		// the following information will be filled in later as artifacts
		// propagate through the pipeline.
		SignedArtifactDir: "",
		TestResultsPath:   "",
		SignedSha256Sum:   "",
	}

	if err := writeInstructions(&pi, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(7)
	}
}

func writeInstructions(pi *archive.PipelineInstructions, outDir string) error {
	instructionDir := filepath.Join(outDir, "instructions")
	if err := os.MkdirAll(instructionDir, 0o700); err != nil {
		return err
	}

	base := fmt.Sprintf("%s-%s-%s.json", pi.Pkg, pi.Distro, pi.SanitizedArch())
	b, err := json.Marshal(pi)
	if err != nil {
		return err
	}

	instructionPath := filepath.Join(instructionDir, base)
	return os.WriteFile(instructionPath, b, 0o600)
}

func findArtifact(artifactDir string) (string, error) {
	var err error
	globs := []string{"*.deb", "*.rpm", "*.zip"}
	matches := []string{}
	for _, glob := range globs {
		m, err := filepath.Glob(filepath.Join(artifactDir, glob))
		if err != nil {
			continue
		}

		matches = append(matches, m...)
	}

	artifact := ""
	err = nil
	switch len(matches) {
	case 0:
		err = fmt.Errorf("no output artifact found")
	case 1:
		artifact = matches[0]
	default:
		err = fmt.Errorf("multiple artifact files found, aborting")
	}

	if err != nil {
		return "", err
	}

	return artifact, nil
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

	target, err := targets.GetTarget(ctx, cfg.Distro, client, platform)
	if err != nil {
		return nil, err
	}
	out := target.Make(cfg)
	return out, nil
}
