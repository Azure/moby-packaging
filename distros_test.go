package main

import (
	"context"
	_ "embed"
	"testing"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
	"github.com/Azure/moby-packaging/targets"
)

const (
	jammy    = "jammy"
	focal    = "focal"
	bionic   = "bionic"
	bullseye = "bullseye"
	buster   = "buster"
	rhel9    = "rhel9"
	rhel8    = "rhel8"
	mariner2 = "mariner2"
)

var distros = map[string]func(context.Context, *testing.T, *dagger.Client) *dagger.Container{
	jammy:    Jammy,
	focal:    Focal,
	bionic:   Bionic,
	bullseye: Bullseye,
	buster:   Buster,
	rhel9:    Rhel9,
	rhel8:    Rhel8,
	mariner2: Mariner2,
}

//go:embed tests/deb/install.sh
var debInstall string

func Jammy(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.JammyRef)
	return apt.Install(c, client.CacheVolume(targets.JammyAptCacheKey), client.CacheVolume(targets.JammyAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}).
		WithNewFile("/opt/moby/install.sh", dagger.ContainerWithNewFileOpts{Contents: debInstall, Permissions: 0744})
}

func Focal(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.FocalRef)
	return apt.Install(c, client.CacheVolume(targets.FocalAptCacheKey), client.CacheVolume(targets.FocalAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"systemctl", "enable", "systemd-udevd"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}).
		WithNewFile("/opt/moby/install.sh", dagger.ContainerWithNewFileOpts{Contents: debInstall, Permissions: 0744})
}

func Bionic(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BionicRef)
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.Install(c2, client.CacheVolume(targets.BionicAptCacheKey), client.CacheVolume(targets.BionicAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Bullseye(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BullseyeRef)
	deb := client.HTTP("https://packages.microsoft.com/config/debian/11/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.Install(c2, client.CacheVolume(targets.BullseyeAptCacheKey), client.CacheVolume(targets.BullseyeAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Buster(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BusterRef)
	deb := client.HTTP("https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.Install(c2, client.CacheVolume(targets.BusterAptCacheKey), client.CacheVolume(targets.BusterAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Rhel9(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.Rhel9Ref)
	c = c.WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
		`})
	return c
}

func Rhel8(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.Rhel8Ref)
	c = c.WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/8/packages-microsoft-prod.rpm
		`})
	return c
}

func Mariner2(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	return client.Container().From(targets.Mariner2Ref)
}
