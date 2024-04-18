package containerd

import (
	_ "embed"

	"github.com/Azure/moby-packaging/pkg/archive"
)

var (
	Archives_2_0 = map[string]archive.Archive{
		"bookworm": DebArchive_2_0,
		"buster":   DebArchive_2_0,
		"bullseye": DebArchive_2_0,
		"bionic":   DebArchive_2_0,
		"focal":    DebArchive_2_0,
		"jammy":    DebArchive_2_0,
		"noble":    DebArchive_2_0,
		"centos7":  RPMArchive_2_0,
		"rhel8":    RPMArchive_2_0,
		"rhel9":    RPMArchive_2_0,
		"mariner2": MarinerArchive_2_0,
		"windows":  BaseArchive_2_0,
	}

	BaseArchive_2_0 = archive.Archive{
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
			"/build/src/bin/containerd-shim-runc-v2",
			"/build/src/bin/ctr",
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

	DebArchive_2_0 = archive.Archive{
		Name:     BaseArchive_2_0.Name,
		Webpage:  BaseArchive_2_0.Webpage,
		Files:    BaseArchive_2_0.Files,
		Systemd:  BaseArchive_2_0.Systemd,
		Binaries: BaseArchive_2_0.Binaries,
		RuntimeDeps: []string{
			"moby-runc (>= 1.0.2)",
		},
		Recommends: []string{
			"ca-certificates",
		},
		Conflicts: []string{
			"containerd", "containerd.io", "moby-engine (<= 3.0.12)",
		},
		Replaces: []string{
			"containerd", "containerd.io",
		},
		Provides: []string{
			"containerd", "containerd.io",
		},
		InstallScripts: []archive.InstallScript{
			{When: archive.PkgActionPostInstall, Script: debPostInstall},
			{When: archive.PkgActionPreRemoval, Script: debPreRm},
			{When: archive.PkgActionPostRemoval, Script: debPostRm},
		},
		Description: BaseArchive_2_0.Description,
	}

	RPMArchive_2_0 = archive.Archive{
		Name:     BaseArchive_2_0.Name,
		Webpage:  BaseArchive_2_0.Webpage,
		Files:    BaseArchive_2_0.Files,
		Systemd:  BaseArchive_2_0.Systemd,
		Binaries: BaseArchive_1_X.Binaries,
		RuntimeDeps: []string{
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
		Conflicts: []string{
			"containerd", "containerd-io", "moby-engine <= 3.0.11",
		},
		InstallScripts: []archive.InstallScript{
			{When: archive.PkgActionPostInstall, Script: rpmPostInstall},
			{When: archive.PkgActionPreRemoval, Script: rpmPreRm},
			{When: archive.PkgActionUpgrade, Script: rpmUpgrade},
		},
		Description: BaseArchive_2_0.Description,
	}

	MarinerArchive_2_0 = func() archive.Archive {
		m := RPMArchive_2_0
		m.RuntimeDeps = []string{
			"/bin/sh",
			"device-mapper-libs >= 1.02.90-1",
			"iptables",
			"libcgroup",
			"libseccomp >= 2.3",
			"moby-containerd >= 1.3.9",
			"moby-runc >= 1.0.2",
			"systemd-units",
			"tar",
			"xz",
		}
		return m
	}()
)
