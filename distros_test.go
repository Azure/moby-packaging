package main

import (
	"github.com/Azure/moby-packaging/pkg/apt"
	"github.com/Azure/moby-packaging/targets"

	"dagger.io/dagger"
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

var distros = map[string]func(*dagger.Client) *dagger.Container{
	jammy:    Jammy,
	focal:    Focal,
	bionic:   Bionic,
	bullseye: Bullseye,
	buster:   Buster,
	rhel9:    Rhel9,
	rhel8:    Rhel8,
	mariner2: Mariner2,
}
var distroIDs = map[string]string{
	jammy:    "ubuntu22.04",
	focal:    "ubuntu20.04",
	bionic:   "ubuntu18.04",
	bullseye: "debian11",
	buster:   "debian10",
	rhel9:    "el9",
	rhel8:    "el8",
	mariner2: "mariner2.0",
}

func Jammy(client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.JammyRef)
	return apt.AptInstall(c, client.CacheVolume(targets.JammyAptCacheKey), client.CacheVolume(targets.JammyAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"systemctl", "enable", "systemd-udevd"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}).
		WithFile("/opt/moby/install.sh", client.Host().Directory("tests/deb").File("install.sh"))
}

func Focal(client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.FocalRef)
	return apt.AptInstall(c, client.CacheVolume(targets.FocalAptCacheKey), client.CacheVolume(targets.FocalAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"systemctl", "enable", "systemd-udevd"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}).
		WithFile("/opt/moby/install.sh", client.Host().Directory("tests/deb").File("install.sh"))
}

func Bionic(client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BionicRef)
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.AptInstall(c2, client.CacheVolume(targets.BionicAptCacheKey), client.CacheVolume(targets.BionicAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Bullseye(client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BullseyeRef)
	deb := client.HTTP("https://packages.microsoft.com/config/debian/11/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.AptInstall(c2, client.CacheVolume(targets.BullseyeAptCacheKey), client.CacheVolume(targets.BullseyeAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Buster(client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.BusterRef)
	deb := client.HTTP("https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb")
	c2 := c.WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
	c2 = apt.AptInstall(c2, client.CacheVolume(targets.BusterAptCacheKey), client.CacheVolume(targets.BusterAptLibCacheKey), "aptly")
	return c.WithRootfs(c2.Rootfs())
}

func Rhel9(client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.Rhel9Ref)
	c = c.WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
		`})
	return c
}

func Rhel8(client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.Rhel8Ref)
	c = c.WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/8/packages-microsoft-prod.rpm
		`})
	return c
}

func Mariner2(client *dagger.Client) *dagger.Container {
	return client.Container().From(targets.Mariner2Ref)
}
