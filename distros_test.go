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
	bookworm = "bookworm"
	buster   = "buster"
	rhel9    = "rhel9"
	rhel8    = "rhel8"
	mariner2 = "mariner2"
)

type DistroTestHelper interface {
	Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container
	Installer(ctx context.Context, client *dagger.Client) *dagger.File
	FormatVersion(version, revision string) string
}

var distros = map[string]DistroTestHelper{
	jammy:    JammyTestHelper{},
	focal:    FocalTestHelper{},
	bionic:   BionicTestHelper{},
	bullseye: BullseyeTestHelper{},
	bookworm: BookwormTestHelper{},
	buster:   BusterTestHelper{},
	rhel9:    Rhel9TestHelper{},
	rhel8:    Rhel8TestHelper{},
	mariner2: Mariner2TestHelper{},
}

var (
	//go:embed tests/deb/install.sh
	debInstall string

	//go:embed tests/centos8/install.sh
	rhel8Install string

	//go:embed tests/mariner2/install.sh
	mariner2Install string
)

type JammyTestHelper struct{}

func (JammyTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	c := client.Container().From(targets.JammyRef)
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb")
	return apt.Install(c, client.CacheVolume(targets.JammyAptCacheKey), client.CacheVolume(targets.JammyAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (JammyTestHelper) FormatVersion(version, revision string) string {
	return version + "-ubuntu22.04u" + revision
}

func (JammyTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

type FocalTestHelper struct{}

func (FocalTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.FocalRef)
	return apt.Install(c, client.CacheVolume(targets.FocalAptCacheKey), client.CacheVolume(targets.FocalAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (FocalTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (FocalTestHelper) FormatVersion(version, revision string) string {
	return version + "-ubuntu20.04u" + revision
}

type BionicTestHelper struct{}

func (BionicTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BionicRef)
	return apt.Install(c, client.CacheVolume(targets.BionicAptCacheKey), client.CacheVolume(targets.BionicAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (BionicTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (BionicTestHelper) FormatVersion(version, revision string) string {
	return version + "-ubuntu18.04u" + revision
}

type BullseyeTestHelper struct{}

func (BullseyeTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/debian/11/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BullseyeRef)
	return apt.Install(c, client.CacheVolume(targets.BullseyeAptCacheKey), client.CacheVolume(targets.BullseyeAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (BullseyeTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (BullseyeTestHelper) FormatVersion(version, revision string) string {
	return version + "-debian11u" + revision
}

type BookwormTestHelper struct{}

func (BookwormTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BullseyeRef)
	return apt.Install(c, client.CacheVolume(targets.BullseyeAptCacheKey), client.CacheVolume(targets.BullseyeAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (BookwormTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (BookwormTestHelper) FormatVersion(version, revision string) string {
	return version + "-debian12u" + revision
}

type BusterTestHelper struct{}

func (BusterTestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	deb := client.HTTP("https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb")

	c := client.Container().From(targets.BusterRef)
	return apt.Install(c, client.CacheVolume(targets.BusterAptCacheKey), client.CacheVolume(targets.BusterAptLibCacheKey),
		"systemd", "strace", "ssh", "udev", "iptables", "jq",
	).
		WithExec([]string{"systemctl", "enable", "ssh"}).
		WithExec([]string{"update-alternatives", "--set", "iptables", "/usr/sbin/iptables-legacy"}).
		WithMountedFile("/tmp/packages-microsoft-prod.deb", deb).
		WithExec([]string{"/usr/bin/dpkg", "-i", "/tmp/packages-microsoft-prod.deb"})
}

func (BusterTestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", debInstall, dagger.DirectoryWithNewFileOpts{Permissions: 0743}).File("install.sh")
}

func (BusterTestHelper) FormatVersion(version, revision string) string {
	return version + "-debian10u" + revision
}

type Rhel9TestHelper struct{}

func (Rhel9TestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	return client.Container().From(targets.Rhel9Ref).
		WithExec([]string{
			"dnf", "install", "-y",
			"createrepo_c", "systemd", "strace", "openssh-server", "openssh-clients", "udev", "iptables", "dnf-command(config-manager)", "jq",
		}).
		WithExec([]string{"systemctl", "enable", "sshd"}).
		WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
		`})
}

func (Rhel9TestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", rhel8Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (Rhel9TestHelper) FormatVersion(version, revision string) string {
	return version + "-" + revision + ".el9"
}

type Rhel8TestHelper struct{}

func (Rhel8TestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
	return client.Container().From(targets.Rhel8Ref).
		WithExec([]string{
			"dnf", "install", "-y",
			"createrepo_c", "systemd", "strace", "openssh-server", "openssh-clients", "udev", "iptables", "dnf-command(config-manager)", "dnf-utils", "util-linux", "jq",
		}).
		WithExec([]string{"systemctl", "enable", "sshd"}).
		WithExec([]string{"bash", "-c", `
			dnf install -y https://packages.microsoft.com/config/rhel/8/packages-microsoft-prod.rpm
		`})
}

func (Rhel8TestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", rhel8Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (Rhel8TestHelper) FormatVersion(version, revision string) string {
	return version + "-" + revision + ".el8"
}

type Mariner2TestHelper struct{}

func (Mariner2TestHelper) Image(ctx context.Context, t *testing.T, client *dagger.Client) *dagger.Container {
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

	return c
}

func (Mariner2TestHelper) Installer(ctx context.Context, client *dagger.Client) *dagger.File {
	return client.Container().Rootfs().WithNewFile("install.sh", mariner2Install, dagger.DirectoryWithNewFileOpts{Permissions: 0744}).File("install.sh")
}

func (Mariner2TestHelper) FormatVersion(version, revision string) string {
	return version + "-" + revision + ".cm2"
}
