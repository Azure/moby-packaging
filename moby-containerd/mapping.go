package containerd

import (
	_ "embed"
	"packaging/pkg/archive"
)

var (
	//go:embed postinstall/deb/postinstall
	debPostInstall string
	//go:embed postinstall/deb/prerm
	debPreRm string
	//go:embed postinstall/deb/postrm
	debPostRm string

	//go:embed postinstall/rpm/postinstall
	rpmPostInstall string
	//go:embed postinstall/rpm/prerm
	rpmPreRm string
	//go:embed postinstall/rpm/upgrade
	rpmUpgrade string

	Archive = archive.Archive{
		Name:    "moby-containerd",
		Webpage: "https://github.com/containerd/containerd",
		Files: []archive.File{
			{Source: "/build/src/bin", Dest: "usr/bin"},
			{Source: "/build/man", Dest: "/usr/share/man"},
			{Source: "/build/legal/LICENSE", Dest: "/usr/share/doc/moby-containerd/LICENSE"},
			{Source: "/build/legal/NOTICE", Dest: "/usr/share/doc/moby-containerd/NOTICE.gz", Compress: true},
		},
		Systemd: []archive.Systemd{
			{Source: "/build/src/containerd.service", Dest: "lib/systemd/system/containerd.service"},
		},
		Binaries: []string{
			"/build/src/bin/containerd",
			"/build/src/bin/containerd-shim",
			"/build/src/bin/containerd-shim-runc-v1",
			"/build/src/bin/containerd-shim-runc-v2",
			"/build/src/bin/containerd-stress",
		},
		RuntimeDeps: archive.PkgKindMap{
			archive.PkgKindRPM: {
				"/bin/sh",
				"container-selinux >= 2:2.95",
				"device-mapper-libs >= 1.02.90-1",
				"iptables",
				"libcgroup",
				"libseccomp >= 2.3",
				"moby-containerd >= 1.3.9",
				"moby-runc >= 1.0.2",
				"systemd-units",
				"tar",
				"xz",
			},
			archive.PkgKindDeb: {"moby-runc (>= 1.0.2)"},
		},
		Recommends: archive.PkgKindMap{
			archive.PkgKindDeb: {"ca-certificates"},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {"containerd", "containerd.io", "moby-engine (<= 3.0.12)"},
			archive.PkgKindRPM: {"containerd", "containerd-io", "moby-engine <= 3.0.11"},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {"containerd", "containerd.io"},
		},
		Provides: archive.PkgKindMap{
			archive.PkgKindDeb: {"containerd", "containerd.io"},
		},
		InstallScripts: archive.PkgInstallMap{
			archive.PkgKindRPM: {
				{When: archive.PkgActionPostInstall, Script: rpmPostInstall},
				{When: archive.PkgActionPreRemoval, Script: rpmPreRm},
				{When: archive.PkgActionUpgrade, Script: rpmUpgrade},
			},
			archive.PkgKindDeb: {
				{When: archive.PkgActionPostInstall, Script: debPostInstall},
				{When: archive.PkgActionPreRemoval, Script: debPreRm},
				{When: archive.PkgActionPostRemoval, Script: debPostRm},
			},
		},
		WinBinaries: []string{
			"/build/src/bin/containerd.exe",
			"/build/src/bin/containerd-shim-runhcs-v1.exe",
			"/build/src/bin/ctr.exe",
		},
		Description: `Industry-standard container runtime
 containerd is an industry-standard container runtime with an emphasis on
 simplicity, robustness and portability. It is available as a daemon for Linux
 and Windows, which can manage the complete container lifecycle of its host
 system: image transfer and storage, container execution and supervision,
 low-level storage and network attachments, etc.
 .
 containerd is designed to be embedded into a larger system, rather than being
 used directly by developers or end-users.`,
	}
	Dirs = []string{}
)
