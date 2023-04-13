package compose

import "packaging/pkg/archive"

var (
	Mapping = map[string]string{
		"src/bin/docker-compose": "/usr/libexec/docker/cli-plugins/docker-compose",
	}
	Mapping2 = []archive.File{
		{
			Source: "/build/src/bin/docker-compose",
			Dest:   "/usr/libexec/docker/cli-plugins/docker-compose",
		},
		{
			Source: "/build/legal/LICENSE",
			Dest:   "/usr/share/doc/moby-compose/LICENSE",
		},
		{
			Source:   "/build/legal/NOTICE",
			Dest:     "/usr/share/doc/moby-compose/NOTICE.gz",
			Compress: true,
		},
	}

	Archive = archive.Archive{
		Name:    "moby-compose",
		Webpage: "https://github.com/docker/compose-cli",
		Files: []archive.File{
			{
				Source: "/build/src/bin/docker-compose",
				Dest:   "/usr/libexec/docker/cli-plugins/docker-compose",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-compose/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-compose/NOTICE.gz",
				Compress: true,
			},
		},
		Systemd:  []archive.Systemd{},
		Postinst: []string{},
		Binaries: []string{
			"/build/src/bin/docker-compose",
		},
		RuntimeDeps: map[archive.PkgKind][]string{
			archive.PkgKindRPM: {
				"/bin/sh",
				"container-selinux >= 2:2.95",
				"device-mapper-libs >= 1.02.90-1",
				"iptables",
				"libcgroup",
				"moby-cli",
				"moby-containerd >= 1.3.9",
				"moby-runc >= 1.0.2",
				"systemd-units",
				"tar",
				"xz",
			},
			archive.PkgKindDeb: {
				"moby-cli",
			},
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"docker-ce",
				"docker-ee",
				"docker-ce-cli",
				"docker-ee-cli",
			},
			archive.PkgKindRPM: {
				"docker-ce",
				"docker-ee",
			},
		},
		Description: `A Docker CLI plugin which allows you to run Docker Compose applications from the Docker CLI.`,
	}
)
