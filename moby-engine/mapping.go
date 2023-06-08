package engine

// #!/usr/bin/dh-exec

// #tini/build/tini-static => /usr/bin/docker-init

import (
	_ "embed"

	"github.com/Azure/moby-packaging/pkg/archive"
)

var (
	//go:embed postinstall/rpm/postinstall
	rpmPostInstall string
	//go:embed postinstall/rpm/prerm
	rpmPreRm string
	//go:embed postinstall/rpm/upgrade
	rpmUpgrade string

	//go:embed postinstall/deb/postinstall
	debPostInstall string
	//go:embed postinstall/deb/prerm
	debPreRm string
	//go:embed postinstall/deb/postrm
	debPostRm string

	Archives = map[string]archive.Archive{
		"buster":   DebArchive,
		"bullseye": DebArchive,
		"bionic":   DebArchive,
		"focal":    DebArchive,
		"centos7":  RPMArchive,
		"rhel8":    RPMArchive,
		"rhel9":    RPMArchive,
		"windows":  BaseArchive,
		"jammy":    DebArchive,
		"mariner2": MarinerArchive,
	}

	BaseArchive = archive.Archive{
		Name:    "moby-engine",
		Webpage: "https://github.com/moby/moby",
		Files: []archive.File{
			{Source: "/build/systemd/docker.socket", Dest: "/lib/systemd/system/docker.socket"},
			{Source: "/build/src/contrib/nuke-graph-directory.sh", Dest: "/usr/share/moby-engine/contrib/nuke-graph-directory.sh"},
			{Source: "/build/src/contrib/check-config.sh", Dest: "/usr/share/moby-engine/contrib/check-config.sh"},
			{Source: "/build/bundles/dynbinary-daemon/dockerd", Dest: "/usr/bin/dockerd"},
			{Source: "/build/src/libnetwork/docker-proxy", Dest: "/usr/bin/docker-proxy"},
			{Source: "/build/src/contrib/udev/80-docker.rules", Dest: "/lib/udev/rules.d/80-moby-engine.rules"},
			{Source: "", Dest: "/etc/docker", IsDir: true},
			{Source: "/build/legal/LICENSE", Dest: "/usr/share/doc/moby-engine/LICENSE"},
			{Source: "/build/legal/NOTICE", Dest: "/usr/share/doc/moby-engine/NOTICE.gz", Compress: true},
		},
		Systemd: []archive.Systemd{
			{Source: "/build/systemd/docker.service", Dest: "/lib/systemd/system/docker.service"},
		},
		Binaries:    []string{"/build/bundles/dynbinary-daemon/dockerd", "/build/src/libnetwork/docker-proxy"},
		WinBinaries: []string{"/build/src/bundles/binary-daemon/dockerd.exe"},
		Description: `Docker container platform (engine package)
  Moby is an open-source project created by Docker to enable and accelerate software containerization.`,
	}

	DebArchive = archive.Archive{
		Name:     BaseArchive.Name,
		Webpage:  BaseArchive.Webpage,
		Files:    BaseArchive.Files,
		Systemd:  BaseArchive.Systemd,
		Binaries: BaseArchive.Binaries,
		RuntimeDeps: []string{
			"moby-containerd (>= 1.4.3)", "moby-runc (>= 1.0.2)", "moby-tini (>= 0.19.0)",
		},
		Recommends: []string{
			"apparmor",
			"ca-certificates",
			"iptables",
			"kmod",
			"moby-cli",
			"pigz",
			"xz-utils",
		},
		Suggests: []string{
			"aufs-tools",
			"cgroupfs-mount | cgroup-lite",
			"git",
		},
		Conflicts: []string{
			"docker",
			"docker-ce",
			"docker-ee",
			"docker-engine",
			"docker-engine-cs",
			"docker.io",
			"lxc-docker",
			"lxc-docker-virtual-package",
		},
		Replaces: []string{
			"docker",
			"docker-ce",
			"docker-ee",
			"docker-engine",
			"docker-engine-cs",
			"docker.io",
			"lxc-docker",
			"lxc-docker-virtual-package",
		},
		InstallScripts: []archive.InstallScript{
			{
				When:   archive.PkgActionPostInstall,
				Script: debPostInstall,
			},
			{
				When:   archive.PkgActionPreRemoval,
				Script: debPreRm,
			},
			{
				When:   archive.PkgActionPostRemoval,
				Script: debPostRm,
			},
		},
		Description: BaseArchive.Description,
	}

	RPMArchive = archive.Archive{
		Name:     BaseArchive.Name,
		Webpage:  BaseArchive.Webpage,
		Files:    BaseArchive.Files,
		Systemd:  BaseArchive.Systemd,
		Binaries: BaseArchive.Binaries,
		RuntimeDeps: []string{
			"/bin/sh",
			"container-selinux >= 2:2.95",
			"device-mapper-libs >= 1.02.90-1",
			"iptables",
			"libcgroup",
			"moby-containerd >= 1.3.9",
			"moby-runc >= 1.0.2",
			"systemd-units",
			"tar",
			"xz",
		},
		Recommends: []string{},
		Suggests:   []string{},
		Conflicts: []string{
			"docker",
			"docker-io",
			"docker-engine-cs",
			"docker-ee",
		},
		Replaces: []string{},
		InstallScripts: []archive.InstallScript{
			{
				When:   archive.PkgActionPostInstall,
				Script: rpmPostInstall,
			},
			{
				When:   archive.PkgActionPreRemoval,
				Script: rpmPreRm,
			},
			{
				When:   archive.PkgActionUpgrade,
				Script: rpmUpgrade,
			},
		},
		Description: BaseArchive.Description,
	}

	MarinerArchive = func() archive.Archive {
		m := RPMArchive

		m.RuntimeDeps = []string{
			"/bin/sh",
			"device-mapper-libs >= 1.02.90-1",
			"iptables",
			"libcgroup",
			"moby-containerd >= 1.3.9",
			"moby-runc >= 1.0.2",
			"systemd-units",
			"tar",
			"xz",
		}

		return m
	}()
)
