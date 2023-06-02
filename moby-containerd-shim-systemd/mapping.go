package shim

import (
	_ "embed"

	"github.com/Azure/moby-packaging/pkg/archive"
)

var (
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
		Name:    "moby-containerd-shim-systemd",
		Webpage: "https://github.com/cpuguy83/containerd-shim-systemd-v1",
		Files: []archive.File{
			{
				Source: "/build/src/bin/containerd-shim-systemd-v1",
				Dest:   "/usr/bin/containerd-shim-systemd-v1",
			},
			{
				Source: "/build/systemd/containerd-shim-systemd-v1.socket",
				Dest:   "/lib/systemd/system/containerd-shim-systemd-v1.socket",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-containerd-shim-systemd/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-containerd-shim-systemd/NOTICE.gz",
				Compress: true,
			},
		},
		Systemd: []archive.Systemd{
			{
				Source: "/build/systemd/containerd-shim-systemd-v1.service",
				Dest:   "/lib/systemd/system/containerd-shim-systemd-v1.service",
			},
		},
		Binaries: []string{
			"/build/src/bin/containerd-shim-systemd-v1",
		},
		Description: `A containerd shim runtime that uses systemd to monitor runc containers`,
	}

	DebArchive = archive.Archive{
		Name:    BaseArchive.Name,
		Webpage: BaseArchive.Webpage,
		Files:   BaseArchive.Files,
		Systemd: BaseArchive.Systemd,
		Binaries: []string{
			"/build/src/bin/containerd-shim-systemd-v1",
		},
		RuntimeDeps: []string{
			"systemd (>= 239)",
			"moby-containerd (>= 1.6)",
		},
		Recommends: []string{
			"moby-runc",
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
		Name:    BaseArchive.Name,
		Webpage: BaseArchive.Webpage,
		Files:   BaseArchive.Files,
		Systemd: BaseArchive.Systemd,
		Binaries: []string{
			"/build/src/bin/containerd-shim-systemd-v1",
		},
		RuntimeDeps: []string{
			"/bin/sh",
			"container-selinux >= 2:2.95",
			"device-mapper-libs >= 1.02.90-1",
			"iptables",
			"libcgroup",
			"moby-containerd >= 1.3.9",
			"moby-containerd >= 1.6, systemd => 239",
			"moby-runc >= 1.0.2",
			"systemd-units",
			"tar",
			"xz",
		},
		InstallScripts: []archive.InstallScript{},
		Description:    BaseArchive.Description,
	}

	MarinerArchive = func() archive.Archive {
		m := RPMArchive
		m.RuntimeDeps = []string{
			"/bin/sh",
			"device-mapper-libs >= 1.02.90-1",
			"iptables",
			"libcgroup",
			"moby-containerd >= 1.3.9",
			"moby-containerd >= 1.6, systemd => 239",
			"moby-runc >= 1.0.2",
			"systemd-units",
			"tar",
			"xz",
		}
		return m
	}()
)
