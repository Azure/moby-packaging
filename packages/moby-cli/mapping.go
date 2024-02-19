package cli

import (
	_ "embed"

	"github.com/Azure/moby-packaging/pkg/archive"
)

// #!/usr/bin/dh-exec

var (
	//go:embed postinstall/deb/postinstall
	debPostInst string

	//go:embed postinstall/rpm/postinstall
	rpmPostInst string

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
		Name:    "moby-cli",
		Webpage: "https://github.com/docker/cli",
		Files: []archive.File{
			{
				Source: "/build/src/build/docker",
				Dest:   "/usr/bin/docker",
			},
			{
				Source: "/build/src/contrib/completion/zsh/_docker",
				Dest:   "/usr/share/zsh/vendor-completions/_docker",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-cli/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-cli/NOTICE.gz",
				Compress: true,
			},
			{
				Source:   "/build/src/contrib/completion/bash/docker",
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
		Description: `Docker container platform (client package)
 Docker is a platform for developers and sysadmins to develop, ship, and run
 applications. Docker lets you quickly assemble applications from components and
 eliminates the friction that can come when shipping code. Docker lets you get
 your code tested and deployed into production as fast as possible.
 .
 This package provides the "docker" client binary (and supporting files).`,
	}

	DebArchive = archive.Archive{
		Name:        BaseArchive.Name,
		Webpage:     BaseArchive.Webpage,
		Files:       BaseArchive.Files,
		Binaries:    []string{"/build/src/build/docker"},
		RuntimeDeps: []string{},
		Recommends: []string{
			"ca-certificates",
			"git",
			"moby-buildx",
			"pigz",
			"xz-utils",
		},
		Suggests: []string{
			"moby-engine",
		},
		Conflicts: []string{
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
		Replaces: []string{
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
		InstallScripts: []archive.InstallScript{
			{
				When:   archive.PkgActionPostInstall,
				Script: debPostInst,
			},
		},
		Description: BaseArchive.Description,
	}

	RPMArchive = archive.Archive{
		Name:     BaseArchive.Name,
		Webpage:  BaseArchive.Webpage,
		Files:    BaseArchive.Files,
		Binaries: []string{"/build/src/build/docker"},
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
		InstallScripts: []archive.InstallScript{
			{
				When:   archive.PkgActionPostInstall,
				Script: rpmPostInst,
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
