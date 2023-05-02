package main

import (
	"bufio"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/targets"
	"github.com/Azure/moby-packaging/testutil"
	"github.com/joshdk/go-junit"
)

const entrypointVersion = "892ed9a42ceb5f9a9c7198adfc316da64a573274"

var (
	//go:embed tests/setup_ssh.service
	setupSSHService string

	//go:embed tests/setup_ssh.sh
	setupSSH string

	//go:embed tests/docker-entrypoint.sh
	entrypointCmd string

	//go:embed tests/test_runner.sh
	testRunnerCmd string

	//go:embed tests/test.sh
	testSH string
)

func testPackage(ctx context.Context, t *testing.T, client *dagger.Client, spec *archive.Spec) {
	// set up the daemon container
	getContainer, ok := distros[spec.Distro]
	if !ok {
		t.Fatalf("unknown distro: %s", spec.Distro)
	}

	buildOutput, err := do(ctx, client.Pipeline("Build "+spec.Pkg+" for testing"), spec)
	if err != nil {
		t.Fatal(err)
	}

	batsCore, batsHelpers := makeBats(client)

	qemu := testutil.NewQemuImg(ctx, client.Pipeline("Qemu"))

	c := getContainer(ctx, t, client.Pipeline("Setup "+spec.Distro+"/"+spec.Arch))

	vmImage := c.Pipeline("Build VM rootfs").
		WithDirectory("/opt/bats", batsCore).
		WithExec([]string{"/bin/sh", "-c", "cd /opt/bats && ./install.sh /usr/local"}).
		WithDirectory("/opt/moby/test_helper", batsHelpers).
		WithNewFile("/opt/moby/test.sh", dagger.ContainerWithNewFileOpts{Contents: testSH, Permissions: 0744}).
		WithDirectory("/lib/modules", qemu.Pipeline("kernel modules").Directory("/lib/modules"))

	goCtr := client.Pipeline("golang").Container().From(targets.GoRef).
		WithMountedCache("/go/pkg/mod", client.CacheVolume(targets.GoModCacheKey)).
		WithEnvVariable("CGO_ENABLED", "0")

	aptly := goCtr.Pipeline("aptly").WithExec([]string{"go", "install", "github.com/aptly-dev/aptly@v1.5.0"}).File("/go/bin/aptly")
	vmImage = vmImage.WithFile("/usr/local/bin/aptly", aptly)

	entrypointBin := goCtr.Pipeline("qemu-micro-env entrypoint").
		WithExec([]string{"go", "install", "github.com/cpuguy83/qemu-micro-env/cmd/entrypoint@" + entrypointVersion}).
		File("/go/bin/entrypoint")

	resolvConf, err := c.WithExec([]string{"cat", "/etc/resolv.conf"}).Stdout(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rootfs := vmImage.WithNewFile("/usr/local/bin/setup_ssh", dagger.ContainerWithNewFileOpts{
		Contents:    setupSSH,
		Permissions: 0744,
	}).
		WithNewFile("/lib/systemd/system/setup_ssh.service", dagger.ContainerWithNewFileOpts{
			Contents:    setupSSHService,
			Permissions: 0644,
		}).
		WithExec([]string{"systemctl", "enable", "setup_ssh.service"}).Rootfs().
		WithNewFile("/etc/resolv.conf", resolvConf)

	qcow := testutil.QcowFromDir(ctx, rootfs, qemu.Pipeline("Build VM qcow2"))

	// Generate a unique ID to store the socket files in
	// This should *not* be shared between builds, hence the unique key.
	buf := make([]byte, 16)
	n, err := rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	sockets := client.CacheVolume("qemu-micro-env-sockets-" + hex.EncodeToString(buf[:n][:12]))

	runner := qemu.Pipeline("Qemu Exec").
		WithMountedFile("/tmp/rootfs-base.qcow2", qcow).
		WithMountedFile("/usr/local/bin/docker-entrypoint", entrypointBin).
		WithMountedCache("/tmp/sockets", sockets).
		WithNewFile("/usr/local/bin/docker-entrypoint.sh", dagger.ContainerWithNewFileOpts{Contents: entrypointCmd, Permissions: 0744}).
		WithExec([]string{"/bin/sh", "-c", "chown -R 65534:65534 /tmp/sockets"}).
		WithEnvVariable("DEBUG", strconv.FormatBool(flDebug)).
		WithExposedPort(22, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp, Description: "VM ssh"}).
		WithExec([]string{"docker-entrypoint.sh"}, dagger.ContainerWithExecOpts{
			InsecureRootCapabilities: true,
		})

	const svc = "testvm"

	testRunner := qemu.Pipeline("Test Runner", dagger.ContainerPipelineOpts{
		Description: "Configure and run tests in the guest VM",
		Labels:      []dagger.PipelineLabel{{Name: "test", Value: "true"}},
	}).
		WithEnvVariable("SSH_HOST", svc).
		WithMountedCache("/tmp/sockets", sockets).
		WithEnvVariable("SSH_AUTH_SOCK", "/tmp/sockets/agent.sock").
		// Set the test package version in the environment so the test runner can use it to install and test by package version
		WithEnvVariable("TEST_EVAL_VARS", pkgTestEnvEval(ctx, t, spec, c)).
		WithMountedDirectory("/tmp/pkg", buildOutput).
		WithNewFile("/usr/local/bin/test_runner.sh", dagger.ContainerWithNewFileOpts{Contents: testRunnerCmd, Permissions: 0774}).
		WithServiceBinding(svc, runner).
		// TODO: It would be really nice if we could move these tests out of bats and into go tests.
		//    Gist of it would be to create a go subtest for each test case and use ssh to run the test.
		//    This would just allow us to more easily integrate with the test framework and get better reporting.
		WithExec([]string{"test_runner.sh"})

	// Now take the test report (which is in junit format).
	// We'll parse that and create a subtest for each test case.
	report := testRunner.Pipeline("Test Report").File("/tmp/report.xml")
	dt, err := report.Contents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	suites, err := junit.Ingest([]byte(dt))
	if err != nil {
		t.Fatal(err)
	}

	// For each test case, create a subtest and check the status.
	// We'll mark the test as failed/skipped accordingly.
	// Unfortunately the run times of these tests will be zeroed out in go since we can't control that from here.
	for _, s := range suites {
		for _, tc := range s.Tests {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				if tc.Error != nil {
					t.Fatal(tc.Error)
				}

				if tc.Status == junit.StatusSkipped {
					t.Skip(tc.Message)
				}

				if s.SystemOut != "" {
					t.Log(s.SystemOut)
				}
			})
		}
	}
}

// package names to git commit hashes to test with
var testPackages = []archive.Spec{
	{
		Pkg:      "moby-runc",
		Repo:     "https://github.com/opencontainers/runc.git",
		Revision: "4",
		Commit:   "5fd4c4d144137e991c4acebb2146ab1483a97925",
	},
	{
		Pkg:      "moby-containerd",
		Repo:     "https://github.com/containerd/containerd.git",
		Revision: "3",
		Commit:   "1fbd70374134b891f97ce19c70b6e50c7b9f4e0d",
	},
	{
		Pkg:      "moby-engine",
		Repo:     "https://github.com/moby/moby.git",
		Revision: "9",
		Commit:   "d7573ab8672555762688f4c7ab8cc69ae8ec1a47",
	},
	{
		Pkg:      "moby-init",
		Repo:     "https://github.com/krallin/tini.git",
		Revision: "7",
		Commit:   "de40ad007797e0dcd8b7126f27bb87401d224240",
	},
	{
		Pkg:      "moby-cli",
		Repo:     "https://github.com/docker/cli.git",
		Revision: "2",
		Commit:   "e92dd87c3209361f29b692ab4b8f0f9248779297",
	},
	{
		Pkg:      "moby-buildx",
		Repo:     "https://github.com/docker/buildx.git",
		Revision: "3",
		Commit:   "00ed17df6d20f3ca4553d45789264cdb78506e5f",
	},
	{
		Pkg:      "moby-compose",
		Repo:     "https://github.com/docker/compose.git",
		Revision: "13",
		Commit:   "00c60da331e7a70af922b1afcce5616c8ab6df36",
	},
}

func TestPackages(t *testing.T) {
	ctx := signalCtx

	client := getClient(ctx, t)

	// If a build spec was provided, only run that.
	if buildSpec != nil {
		t.Run(filepath.Join(buildSpec.Pkg+"/"+buildSpec.Distro+"/"+buildSpec.Arch), func(t *testing.T) {
			testPackage(ctx, t, client, buildSpec)
		})
		return
	}

	for distro := range distros {
		distro := distro
		t.Run(distro, func(t *testing.T) {
			t.Parallel()
			for _, pkg := range testPackages {
				pkg := pkg
				pkg.Distro = distro

				// Set the tag to a very large number so that we can ensure this
				// is the one that the package manager will install instead of
				// the one from the distro repos.
				pkg.Tag = "99.99.99"

				t.Run(pkg.Pkg, func(t *testing.T) {
					t.Parallel()
					testPackage(ctx, t, client.Pipeline(t.Name()), &pkg)
				})
			}
		})
	}
}

func makeBats(client *dagger.Client) (core *dagger.Directory, helpers *dagger.Directory) {
	client = client.Pipeline("Bats")

	const batsCoreRef = "743b02b27c888eba6bb60931656cc16bd751e544"
	core = client.Git("https://github.com/bats-core/bats-core.git").Commit(batsCoreRef).Tree()

	const batsSupportRef = "24a72e14349690bcbf7c151b9d2d1cdd32d36eb1"
	support := client.Git("https://github.com/bats-core/bats-support.git").Commit(batsSupportRef).Tree()

	const batsAssertRef = "0a8dd57e2cc6d4cc064b1ed6b4e79b9f7fee096f"
	assert := client.Git("https://github.com/bats-core/bats-assert.git").Commit(batsAssertRef).Tree()

	helpers = client.Directory().
		WithDirectory("bats-support", support).
		WithDirectory("bats-assert", assert)
	return core, helpers
}

// getPkgTargetID gets the version ID of the distro that the package has set on it
// This is taken from /etc/os-release, specifically the concatenated ID and the VERSION_ID fields.
func getPkgTargetID(ctx context.Context, t *testing.T, c *dagger.Container) string {
	dt, err := c.File("/etc/os-release").Contents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(strings.NewReader(dt))
	var (
		id      string
		version string
	)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			t.Fatalf("unexpected line in /etc/os-release: %s", line)
		}

		switch key {
		case "ID":
			id, err = strconv.Unquote(val)
			if err != nil {
				if err != strconv.ErrSyntax {
					t.Fatal(err)
				}
				id = val
			}
		case "VERSION_ID":
			version, err = strconv.Unquote(val)
			if err != nil {
				if err != nil {
					if err != strconv.ErrSyntax {
						t.Fatal(err)
					}
					version = val
				}
			}
		}
	}

	if id == "" || version == "" {
		t.Fatalf("missing ID or VERSION_ID in /etc/os-release: %s", dt)
	}
	return id + version
}

// pkgTestEnvEval generates environment variables (or rather a shell script that can be sourced/eval'd to set them) used by the bats package tests
// as the expected package version/commit hash/etc to be installed.
func pkgTestEnvEval(ctx context.Context, t *testing.T, spec *archive.Spec, c *dagger.Container) string {
	// package name should be moby-<pkg>
	_, pkg, ok := strings.Cut(spec.Pkg, "-")
	if !ok {
		t.Fatalf("unexpected package name: %s", pkg)
	}

	pkg = strings.ToUpper(pkg)

	versionID := getPkgTargetID(ctx, t, c)

	b := &strings.Builder{}

	writeVar := func(s string) {
		if _, err := b.WriteString(s); err != nil {
			t.Fatal(err)
		}
		if _, err := b.WriteString(" "); err != nil {
			t.Fatal(err)
		}
	}

	// This is used to both install the specific package version as well as check the package version in the tests
	v := fmt.Sprintf(`TEST_%s_PACKAGE_VERSION="%s+azure-%su%s"`, pkg, spec.Tag, versionID, spec.Revision)
	writeVar(v)

	// This makes it to the tests can check the git commit set on the binary itself
	v = fmt.Sprintf(`TEST_%s_COMMIT="%s"`, pkg, spec.Commit)
	writeVar(v)

	// This is used to check the version reported by the binary
	v = fmt.Sprintf(`TEST_%s_VERSION="%s+azure-%s"`, pkg, spec.Tag, spec.Revision)
	writeVar(v)

	return b.String()
}
