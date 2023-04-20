package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/targets"
	"github.com/Azure/moby-packaging/testutil"
)

var (
	GoVersion = "1.19.5"
	GoRef     = path.Join("mcr.microsoft.com/oss/go/microsoft/golang:" + GoVersion)
)

//go:embed tests/setup_ssh.service
var setupSSHService string

//go:embed tests/setup_ssh.sh
var setupSSH string

//go:embed tests/docker-entrypoint.sh
var entrypointCmd string

const entrypointVersion = "5ebaa181f866e32e59e37face813cf25b74e8911"

//go:embed tests/test_runner.sh
var testRunnerCmd string

//go:embed tests/test.sh
var testSH string

func testPackage(ctx context.Context, t *testing.T, client *dagger.Client, spec *archive.Spec) {
	// set up the daemon container
	getContainer, ok := distros[spec.Distro]
	if !ok {
		t.Fatalf("unknown distro: %s", spec.Distro)
	}

	buildOutput, err := do(ctx, client, spec)
	if err != nil {
		t.Fatal(err)
	}

	batsCore, batsHelpers := makeBats(client)

	qemu := testutil.NewQemuImg(ctx, client)

	c := getContainer(ctx, t, client).
		WithDirectory("/opt/bats", batsCore).
		WithExec([]string{"/bin/sh", "-c", "cd /opt/bats && ./install.sh /usr/local"}).
		WithDirectory("/opt/moby/test_helper", batsHelpers).
		WithNewFile("/opt/moby/test.sh", dagger.ContainerWithNewFileOpts{Contents: testSH, Permissions: 0744}).

	goCtr, err := targets.InstallGo(ctx, c, client.CacheVolume(targets.GoModCacheKey), client.CacheVolume("jammy-go-build-cache-"+spec.Arch))
	if err != nil {
		t.Fatal(err)
	}
	goCtr = goCtr.WithEnvVariable("CGO_ENABLED", "0")

	aptly := goCtr.WithExec([]string{"go", "install", "github.com/aptly-dev/aptly@v1.5.0"}).File("/go/bin/aptly")
	c = c.WithFile("/usr/local/bin/aptly", aptly)

	entrypointBin := goCtr.
		WithExec([]string{"go", "install", "github.com/cpuguy83/qemu-micro-env/cmd/entrypoint@" + entrypointVersion}).
		File("/go/bin/entrypoint")

	resolvConf, err := c.WithExec([]string{"cat", "/etc/resolv.conf"}).Stdout(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rootfs := c.WithNewFile("/usr/local/bin/setup_ssh", dagger.ContainerWithNewFileOpts{
		Contents:    setupSSH,
		Permissions: 0744,
	}).
		WithNewFile("/lib/systemd/system/setup_ssh.service", dagger.ContainerWithNewFileOpts{
			Contents:    setupSSHService,
			Permissions: 0644,
		}).
		WithExec([]string{"systemctl", "enable", "setup_ssh.service"}).Rootfs().
		WithNewFile("/etc/resolv.conf", resolvConf)

	qcow := testutil.QcowFromDir(ctx, rootfs, qemu)

	buf := make([]byte, 16)
	n, err := rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	sockets := client.CacheVolume("qemu-micro-env-sockets-" + hex.EncodeToString(buf[:n]))

	runner := qemu.
		WithMountedFile("/tmp/rootfs-base.qcow2", qcow).
		WithMountedFile("/usr/local/bin/docker-entrypoint", entrypointBin).
		WithMountedCache("/tmp/sockets", sockets).
		WithNewFile("/usr/local/bin/docker-entrypoint.sh", dagger.ContainerWithNewFileOpts{Contents: entrypointCmd, Permissions: 0744}).
		WithExec([]string{"/bin/sh", "-c", "chown -R 65534:65534 /tmp/sockets"}).
		WithEnvVariable("DEBUG", strconv.FormatBool(flDebug)).
		WithExec([]string{"docker-entrypoint.sh"}, dagger.ContainerWithExecOpts{
			InsecureRootCapabilities: true,
		}).
		WithExposedPort(22, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp, Description: "VM ssh"})

	const svc = "testvm"

	testRunner := qemu.
		WithEnvVariable("SSH_HOST", svc).
		WithMountedCache("/tmp/sockets", sockets).
		WithEnvVariable("SSH_AUTH_SOCK", "/tmp/sockets/agent.sock").
		WithMountedDirectory("/tmp/pkg", buildOutput).
		WithServiceBinding(svc, runner).
		WithNewFile("/usr/local/bin/test_runner.sh", dagger.ContainerWithNewFileOpts{Contents: testRunnerCmd, Permissions: 0744}).
		// TODO: It would be really nice if we could move these tests out of bats and into go tests.
		//    Gist of it would be to create a go subtest for each test case and use ssh to run the test.
		//    This would just allow us to more easily integrate with the test framework and get better reporting.
		WithExec([]string{"test_runner.sh"})

	report := testRunner.File("/tmp/report.xml")
	_, err = report.Export(ctx, "_output/report.xml")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackages(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	client := getClient(ctx, t)

	t.Run(filepath.Join(buildSpec.Pkg+"/"+buildSpec.Distro+"/"+buildSpec.Arch), func(t *testing.T) {
		testPackage(ctx, t, client, buildSpec)
	})

}

func makeBats(client *dagger.Client) (core *dagger.Directory, helpers *dagger.Directory) {
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
