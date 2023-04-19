package buildx

import "github.com/Azure/moby-packaging/pkg/archive"

var (
	Mapping = map[string]string{
		"src/docker-buildx": "/usr/libexec/docker/cli-plugins/docker-buildx",
	}
	Mapping2 = []archive.File{
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
	}

	Archive = archive.Archive{
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
		RuntimeDeps: archive.PkgKindMap{
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
		Name:    "moby-buildx",
		Webpage: "https://github.com/docker/buildx",
		Recommends: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"moby-cli",
			},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker-ce",
				"docker-ee",
			},
			archive.PkgKindRPM: {
				"docker-ce",
				"docker-ee",
			},
		},
		Description: `A Docker CLI plugin for extended build capabilities with BuildKit`,
	}
)
