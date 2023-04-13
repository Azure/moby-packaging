package runc

import "packaging/pkg/archive"

var (
	Archive = archive.Archive{
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
		RuntimeDeps: map[archive.PkgKind][]string{
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
		},
		Suggests: archive.PkgKindMap{
			archive.PkgKindDeb: {"moby-containerd"},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {"runc", "moby-engine (<= 3.0.10)"},
			archive.PkgKindRPM: {
				"runc",
				"runc-io",
			},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {"runc"},
		},
		Provides: archive.PkgKindMap{
			archive.PkgKindDeb: {"runc"},
		},
		Description: `CLI tool for spawning and running containers according to the OCI specification
  runc is a CLI tool for spawning and running containers according to the OCI
  specification.`,
	}
)
