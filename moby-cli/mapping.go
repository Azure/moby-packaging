package cli

import (
	_ "embed"
	"packaging/pkg/archive"
)

// #!/usr/bin/dh-exec

var (
	//go:embed postinstall/deb/postinstall
	debPostInst string

	//go:embed postinstall/rpm/postinstall
	rpmPostInst string

	Mapping = map[string]string{
		"src/build":                          "/usr/bin",
		"src/contrib/completion/zsh/_docker": "/usr/share/zsh/vendor-completions/_docker",
		"src/man/man1":                       "/usr/share/man/man1",
		"src/man/man8":                       "/usr/share/man/man8",
	}
	Mapping2 = []archive.File{}
	Archive  = archive.Archive{
		Name:    "moby-cli",
		Webpage: "https://github.com/docker/cli",
		Files: []archive.File{
			{
				Source: "src/build/docker",
				Dest:   "/usr/bin/docker",
			},
			{
				Source: "src/contrib/completion/zsh/_docker",
				Dest:   "/usr/share/zsh/vendor-completions/_docker",
			},
			{
				Source: "debian/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-cli/LICENSE",
			},
			{
				Source:   "debian/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-cli/NOTICE.gz",
				Compress: true,
			},
			{
				Source:   "src/contrib/completion/bash/docker",
				Dest:     "/usr/share/bash-completion/completions/docker",
				Compress: true,
			},
		},
		Systemd: []archive.Systemd{},
		Postinst: []string{
			"/build/debian/moby-cli.postinst",
		},
		Binaries:    []string{"/build/src/build/docker"},
		WinBinaries: []string{"/build/src/build/docker.exe"},
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
		},
		Recommends: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"ca-certificates",
				"git",
				"moby-buildx",
				"pigz",
				"xz-utils",
			},
		},
		Suggests: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"moby-engine",
			},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker",
				"docker-ce",
				"docker-ce-cli",
				"docker-ee",
				"docker-ee-cli",
				"docker-engine",
				"docker-engine-cs",
				"docker.io",
				"lxc-docker",
				"lxc-docker-virtual-package",
			},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker",
				"docker-ce",
				"docker-ce-cli",
				"docker-ee",
				"docker-ee-cli",
				"docker-engine",
				"docker-engine-cs",
				"docker.io",
				"lxc-docker",
				"lxc-docker-virtual-package",
			},
		},
		InstallScripts: archive.PkgInstallMap{
			archive.PkgKindDeb: {
				{
					When:   archive.PkgActionPostInstall,
					Script: debPostInst,
				},
			},
			archive.PkgKindRPM: {
				{
					When:   archive.PkgActionPostInstall,
					Script: rpmPostInst,
				},
			},
		},
		Description: `Docker container platform (client package)
 Docker is a platform for developers and sysadmins to develop, ship, and run
 applications. Docker lets you quickly assemble applications from components and
 eliminates the friction that can come when shipping code. Docker lets you get
 your code tested and deployed into production as fast as possible.
 .
 This package provides the "docker" client binary (and supporting files).`,
	}
)
