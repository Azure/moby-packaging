package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"os"
	"os/signal"
	"packaging/targets"
	"packaging/testutil"
	"path"
	"testing"

	"dagger.io/dagger"
)

var (
	GoVersion = "1.19.5"
	GoRef     = path.Join("mcr.microsoft.com/oss/go/microsoft/golang:" + GoVersion)
)

const setupSSHService = `
[Unit]
Description=Setup SSH Auth for VM

[Service]
Type=oneshot
ExecStart=/usr/local/bin/setup_ssh
RemainAfterExit=yes

# Set the standard input of our service to the fifo created by qemu
StandardInput=file:/dev/virtio-ports/authorized_keys

[Install]
WantedBy=multi-user.target
`

const setupSSH = `#!/bin/sh
mkdir -p /root/.ssh
while read -r line; do
	if [ -z "$line" ]; then
		continue
	fi
	echo "$line" >> /root/.ssh/authorized_keys
	break
done
chmod 0600 /root/.ssh/authorized_keys
`

const entrypointCmd = `
if [ ! -c /dev/kvm ]; then
	mknod /dev/kvm c 10 232
	chmod a+rw /dev/kvm
fi
rm /tmp/rootfs.qcow2
qemu-img create -f qcow2 -b /tmp/rootfs-base.qcow2 -F qcow2 /tmp/rootfs.qcow2
exec /usr/local/bin/docker-entrypoint --vm-port-forward=22 --vm-port-forward=8080 --uid=65534 --gid=65534
`

const entrypointVersion = "5ebaa181f866e32e59e37face813cf25b74e8911"

func TestPackage(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// set up the daemon container
	getContainer, ok := distros[buildSpec.Distro]
	if !ok {
		t.Fatalf("unknown distro: %s", buildSpec.Distro)
	}

	client := getClient(ctx, t)
	buildOutput, err := do(ctx, client, buildSpec)
	if err != nil {
		t.Fatal(err)
	}

	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		t.Fatal(err)
	}

	batsCore, batsHelpers := makeBats(client)

	c := getContainer(client).
		WithDirectory("/opt/bats", batsCore).
		WithExec([]string{"/bin/sh", "-c", "cd /opt/bats && ./install.sh /usr/local"}).
		WithDirectory("/opt/moby/test_helper", batsHelpers).
		WithFile("/opt/moby/test.sh", client.Host().Directory("tests").File("test.sh"))

	qemu := testutil.NewQemuImg(ctx, client, platform)

	goCtr, err := targets.InstallGo(ctx, c, client.CacheVolume(targets.GoModCacheKey), client.CacheVolume("jammy-go-build-cache-"+string(platform)))
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

	sockets := client.CacheVolume("qemu-micro-env-sockets-" + t.Name() + hex.EncodeToString(buf[:n]))
	runner := qemu.
		WithMountedFile("/tmp/rootfs-base.qcow2", qcow).
		WithMountedFile("/usr/local/bin/docker-entrypoint", entrypointBin).
		WithMountedCache("/tmp/sockets", sockets).
		WithExec([]string{"/bin/sh", "-c", "mkdir -p /tmp/sockets && chown 65534:65534 /tmp/sockets"}).
		WithExec([]string{"/bin/sh", "-c", entrypointCmd}, dagger.ContainerWithExecOpts{
			InsecureRootCapabilities: true,
		}).
		WithExposedPort(22, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp, Description: "VM ssh"})

	const (
		svc      = "testvm"
		apltySvc = "localapt"
	)
	testRunner := qemu.WithServiceBinding(svc, runner).
		WithEnvVariable("SSH_HOST", svc).
		WithMountedCache("/tmp/sockets", sockets).
		WithEnvVariable("SSH_AUTH_SOCK", "/tmp/sockets/agent.sock").
		WithMountedDirectory("/tmp/pkg", buildOutput)

	// TODO: It would be really nice if we could move these tests out of bats and into go tests.
	//    Gist of it would be to create a go subtest for each test case and use ssh to run the test.
	//    This would just allow us to more easily integrate with the test framework and get better reporting.
	testRunner = testRunner.WithExec([]string{
		"/bin/bash",
		"-c",
		`
until [ -S /tmp/sockets/agent.sock ]; do
	echo waiting for ssh agent socket
	sleep 1
done

sshCmd() {
	ssh -o StrictHostKeyChecking=no ${SSH_HOST} $@
}

scpCmd() {
	scp -o StrictHostKeyChecking=no $@
}

scpCmd -r /tmp/pkg ${SSH_HOST}:/var/pkg || exit

sshCmd '/opt/moby/install.sh; let ec=$?; if [ $ec -ne 0 ]; then journalctl -u docker.service; fi; exit $ec' || exit

sshCmd 'bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh'
let ec=$?

set -e
scpCmd ${SSH_HOST}:/opt/moby/TestReport-test.sh.xml /tmp/report.xml

exit $ec
`,
	})

	report := testRunner.File("/tmp/report.xml")
	_, err = report.Export(ctx, "_output/report.xml")
	if err != nil {
		t.Fatal(err)
	}
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
