package shim

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

	Mapping = map[string]string{
		"src/bin/containerd-shim-systemd-v1":        "usr/bin/containerd-shim-systemd-v1",
		"debian/containerd-shim-systemd-v1.service": "lib/systemd/system/containerd-shim-systemd-v1.service",
		"debian/containerd-shim-systemd-v1.socket":  "lib/systemd/system/containerd-shim-systemd-v1.socket",
	}
	Mapping2 = []archive.File{}
	Archive  = archive.Archive{
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
		Postinst: []string{},
		Binaries: []string{
			"/build/src/bin/containerd-shim-systemd-v1",
		},
		RuntimeDeps: map[archive.PkgKind][]string{
			archive.PkgKindRPM: {
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
			archive.PkgKindDeb: {
				"systemd (>= 239)",
				"moby-containerd (>= 1.6)",
			},
		},
		Recommends: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"moby-runc",
			},
		},
		InstallScripts: archive.PkgInstallMap{
			archive.PkgKindDeb: {
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
		},
		Description: `A containerd shim runtime that uses systemd to monitor runc containers`,
	}
)
