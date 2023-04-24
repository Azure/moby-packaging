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

	Mapping2 = []archive.File{}
	Archive  = archive.Archive{
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
		Postinst:    []string{"/build/debian/moby-engine.postinst"},
		Binaries:    []string{"/build/bundles/dynbinary-daemon/dockerd", "/build/src/libnetwork/docker-proxy"},
		WinBinaries: []string{"/build/src/bundles/binary-daemon/dockerd.exe"},
		RuntimeDeps: map[archive.PkgKind][]string{
			archive.PkgKindRPM: {
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
			archive.PkgKindDeb: {"moby-containerd (>= 1.4.3)", "moby-runc (>= 1.0.2)", "moby-init (>= 0.19.0)"},
		},
		Recommends: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"apparmor",
				"ca-certificates",
				"iptables",
				"kmod",
				"moby-cli",
				"pigz",
				"xz-utils",
			},
		},
		Suggests: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"aufs-tools",
				"cgroupfs-mount | cgroup-lite",
				"git",
			},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker",
				"docker-ce",
				"docker-ee",
				"docker-engine",
				"docker-engine-cs",
				"docker.io",
				"lxc-docker",
				"lxc-docker-virtual-package",
			},
			archive.PkgKindRPM: {
				"docker",
				"docker-io",
				"docker-engine-cs",
				"docker-ee",
			},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker",
				"docker-ce",
				"docker-ee",
				"docker-engine",
				"docker-engine-cs",
				"docker.io",
				"lxc-docker",
				"lxc-docker-virtual-package",
			},
		},
		InstallScripts: archive.PkgInstallMap{
			archive.PkgKindRPM: {
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
		Description: `Docker container platform (engine package)
  Moby is an open-source project created by Docker to enable and accelerate software containerization.`,
	}
)
