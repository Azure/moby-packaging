package buildx

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
		"noble":    DebArchive,
		"mariner2": MarinerArchive,
	}

	BaseArchive = archive.Archive{
		Name:        "moby-buildx",
		Webpage:     "https://github.com/docker/buildx",
		Description: `A Docker CLI plugin for extended build capabilities with BuildKit`,
		Files: []archive.File{
			{
				Source: "/build/src/docker-buildx",
				Dest:   "/usr/libexec/docker/cli-plugins/docker-buildx",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-buildx/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-buildx/NOTICE.gz",
				Compress: true,
			},
		},
		Binaries: []string{
			"/build/src/docker-buildx",
		},
	}

	DebArchive = archive.Archive{
		Name:        BaseArchive.Name,
		Webpage:     BaseArchive.Webpage,
		Description: BaseArchive.Description,
		Files:       BaseArchive.Files,
		Binaries: []string{
			"/build/src/docker-buildx",
		},
		Recommends: []string{
			"moby-cli",
		},
		Conflicts: []string{
			"docker-ce",
			"docker-ee",
			"docker-buildx-plugin",
		},
		Replaces: []string{
			"docker-buildx-plugin",
		},
	}

	RPMArchive = archive.Archive{
		Name:        BaseArchive.Name,
		Webpage:     BaseArchive.Webpage,
		Description: BaseArchive.Description,
		Files:       BaseArchive.Files,
		Binaries: []string{
			"/build/src/docker-buildx",
		},
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
		Conflicts: []string{
			"docker-ce",
			"docker-ee",
		},
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
