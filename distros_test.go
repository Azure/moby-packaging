package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"path/filepath"
	"strings"
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

var distros = map[string]func(context.Context, *testing.T, *dagger.Client) (ctr *dagger.Container, pkgInstaller *dagger.File){
	jammy:    Jammy,
	focal:    Focal,
	bionic:   Bionic,
	bullseye: Bullseye,
	buster:   Buster,
	rhel9:    Rhel9,
	rhel8:    Rhel8,
	mariner2: Mariner2,
}

var (
	//go:embed tests/deb/install.sh
	debInstall string

	//go:embed tests/centos8/install.sh
	rhel8Install string

	//go:embed tests/mariner2/install.sh
	mariner2Install string
)

func Jammy(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.JammyRef)

	return apt.Install(c, client.CacheVolume(targets.JammyAptCacheKey), client.CacheVolume(targets.JammyAptLibCacheKey),
			"systemd", "strace", "ssh", "udev", "iptables", "jq",
		).
			WithExec([]string{"systemctl", "enable", "ssh"}).
			WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
			WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
			WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}),
		client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Focal(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.FocalRef)
	return apt.Install(c, client.CacheVolume(targets.FocalAptCacheKey), client.CacheVolume(targets.FocalAptLibCacheKey),
			"systemd", "strace", "ssh", "udev", "iptables", "jq",
		).
			WithExec([]string{"systemctl", "enable", "ssh"}).
			WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
			WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
			WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}),
		client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Bionic(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BionicRef)
	return apt.Install(c, client.CacheVolume(targets.BionicAptCacheKey), client.CacheVolume(targets.BionicAptLibCacheKey),
			"systemd", "strace", "ssh", "udev", "iptables", "jq",
		).
			WithExec([]string{"systemctl", "enable", "ssh"}).
			WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
			WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
			WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}),
		client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Bullseye(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	deb := client.HTTP("https://packages.microsoft.com/config/debian/11/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BullseyeRef)
	return apt.Install(c, client.CacheVolume(targets.BullseyeAptCacheKey), client.CacheVolume(targets.BullseyeAptLibCacheKey),
			"systemd", "strace", "ssh", "udev", "iptables", "jq",
		).
			WithExec([]string{"systemctl", "enable", "ssh"}).
			WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
			WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
			WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}),
		client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Buster(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	deb := client.HTTP("https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BusterRef)
	return apt.Install(c, client.CacheVolume(targets.BusterAptCacheKey), client.CacheVolume(targets.BusterAptLibCacheKey),
			"systemd", "strace", "ssh", "udev", "iptables", "jq",
		).
			WithExec([]string{"systemctl", "enable", "ssh"}).
			WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
			WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
			WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"}),
		client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Rhel9(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	return client.Container().From(targets.Rhel9Ref).
			WithExec([]string{
				"dnf", "install", "-y",
				"createrepo_c", "systemd", "strace", "openssh-server", "openssh-clients", "udev", "iptables", "dnf-command(config-manager)", "jq",
			}).
			WithExec([]string{"systemctl", "enable", "sshd"}).
			WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
		`}),
		client.Container().Rootfs().WithNewFile("install.sh", rhel8Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Rhel8(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	return client.Container().From(targets.Rhel8Ref).
			WithExec([]string{
				"dnf", "install", "-y",
				"createrepo_c", "systemd", "strace", "openssh-server", "openssh-clients", "udev", "iptables", "dnf-command(config-manager)", "dnf-utils", "util-linux", "jq",
			}).
			WithExec([]string{"systemctl", "enable", "sshd"}).
			WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/8/packages-microsoft-prod.rpm
		`}),
		client.Container().Rootfs().WithNewFile("install.sh", rhel8Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func Mariner2(ctx context.Context, t *testing.T, client *dagger.Client) (*dagger.Container, *dagger.File) {
	c := client.Container().From(targets.Mariner2Ref).
		WithExec([]string{
			"tdnf", "install", "-y",
			"createrepo_c", "systemd", "strace", "openssh-server", "openssh-clients", "udev", "iptables", "dnf-command(config-manager)", "dnf-utils", "util-linux", "jq",
		}).
		WithExec([]string{"systemctl", "enable", "sshd"}).
		WithExec([]string{"sed", "-i", "s/PermitRootLogin no/PermitRootLogin yes/", "/etc/ssh/sshd_config"})

	clientPkg := client.Pipeline("Fetch extra mariner packages")
	p, err := clientPkg.DefaultPlatform(ctx)
	if err != nil {
		t.Fatalf("failed to get default platform: %+v", err)
	}
	_, arch, ok := strings.Cut(string(p), "/")
	if !ok {
		t.Fatalf("failed to get arch from platform: %s", p)
	}

	// Mariner provides most of our packages but does not currently provide compose, so pull this in ourselves.
	// TODO: The artifacts API here has a bug where we can't just use the `mariner2/moby-compose/latest/<arch>` endpoint (it gives the wrong package!).
	// 	So we have to use the `mariner2/moby-compose/latest.json` endpoint and parse the JSON to find the right package.
	//  Also, mariner is adding compose to the repo soon, so we should remove this once that happens.
	dt, err := clientPkg.HTTP("https://mobyartifacts.azureedge.net/index/mariner2/moby-compose/latest.json").Contents(ctx)
	if err != nil {
		t.Fatalf("failed to get compose url: %+v", err)
	}

	var pkgs []struct {
		Uri  string `json:"uri"`
		Arch string `json:"arch"`
	}

	if err := json.Unmarshal([]byte(dt), &pkgs); err != nil {
		t.Fatalf("failed to unmarshal compose url: %+v", err)
	}

	for _, pkg := range pkgs {
		if pkg.Arch == arch {
			c = c.WithFile("/var/pkg/"+filepath.Base(pkg.Uri), clientPkg.HTTP(pkg.Uri), dagger.ContainerWithFileOpts{Permissions: 0640})
		}
	}

	return c, client.Container().Rootfs().WithNewFile("install.sh", mariner2Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}
