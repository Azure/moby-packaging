package runc

import "github.com/Azure/moby-packaging/pkg/archive"

var (
	Archives = map[string]archive.Archive{
		"bookworm": DebArchive,
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
		Name:    "moby-runc",
		Webpage: "https://github.com/opencontainers/runc",
		Files: []archive.File{
			{Source: "/build/src/runc", Dest: "/usr/bin/runc"},
			{Source: "/build/src/contrib/completions/bash/runc", Dest: "/usr/share/bash-completion/completions/runc"},
			{Source: "/build/man", Dest: "/usr/share/man", IsDir: true, Compress: true},
			{Source: "/build/legal/LICENSE", Dest: "/usr/share/doc/moby-runc/LICENSE"},
			{Source: "/build/legal/NOTICE", Dest: "/usr/share/doc/moby-runc/NOTICE.gz", Compress: true},
		},
		Systemd:  []archive.Systemd{},
		Postinst: []string{},
		Binaries: []string{"/build/src/runc"},
		Description: `CLI tool for spawning and running containers according to the OCI specification
  runc is a CLI tool for spawning and running containers according to the OCI
  specification.`,
	}

	DebArchive = archive.Archive{
		Name:        BaseArchive.Name,
		Webpage:     BaseArchive.Webpage,
		Files:       BaseArchive.Files,
		Binaries:    []string{"/build/src/runc"},
		RuntimeDeps: []string{},
		Suggests: []string{
			"moby-containerd",
		},
		Conflicts: []string{
			"runc",
			"moby-engine (<= 3.0.10)",
		},
		Replaces: []string{
			"runc",
		},
		Provides: []string{
			"runc",
		},
		Description: BaseArchive.Description,
	}

	RPMArchive = archive.Archive{
		Name:     BaseArchive.Name,
		Webpage:  BaseArchive.Webpage,
		Files:    BaseArchive.Files,
		Binaries: []string{"/build/src/runc"},
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
			"runc",
			"runc-io",
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
