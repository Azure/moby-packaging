package testutil

import (
	"context"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
	"github.com/Azure/moby-packaging/targets"
)

func NewQemuImg(ctx context.Context, client *dagger.Client) *dagger.Container {
	ctr := client.Container().From(targets.JammyRef)

	return apt.Install(
		ctr,
		client.CacheVolume(targets.JammyAptCacheKey),
		client.CacheVolume(targets.JammyAptLibCacheKey),
		"qemu", "qemu-system", "qemu-utils", "openssh-client", "iptables", "linux-image-5.15.*-kvm", "linux-modules-5.15.*-kvm")
}

// QcowFromDir creates a qcow2 image from a dagger directory.
func QcowFromDir(ctx context.Context, dir *dagger.Directory, qemuCtr *dagger.Container) *dagger.File {
	return qemuCtr.
		WithMountedDirectory("/tmp/rootfs", dir).
		WithExec([]string{"/bin/sh", "-c", `
		truncate -s 10G /tmp/rootfs.img
		mkfs.ext4 -d /tmp/rootfs /tmp/rootfs.img
		qemu-img convert /tmp/rootfs.img -O qcow2 /tmp/rootfs.qcow2
		rm -f /tmp/rootfs.img
		`}).File("/tmp/rootfs.qcow2")
}
