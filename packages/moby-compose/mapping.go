package compose

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
		Description: `A Docker CLI plugin which allows you to run Docker Compose applications from the Docker CLI.`,
	}

	DebArchive = archive.Archive{
		Name:        BaseArchive.Name,
		Webpage:     BaseArchive.Webpage,
		Files:       BaseArchive.Files,
		Description: BaseArchive.Description,
		Binaries: []string{
			"/build/src/bin/docker-compose",
		},
		RuntimeDeps: []string{
			"moby-cli",
		},
		Conflicts: []string{
			"docker-ce",
			"docker-ee",
			"docker-ce-cli",
			"docker-ee-cli",
		},
	}

	RPMArchive = archive.Archive{
		Name:    BaseArchive.Name,
		Webpage: BaseArchive.Webpage,
		Files:   BaseArchive.Files,
		Binaries: []string{
			"/build/src/bin/docker-compose",
		},
		RuntimeDeps: []string{
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
		Conflicts: []string{
			"docker-ce",
			"docker-ee",
		},
		Description: BaseArchive.Description,
	}

	MarinerArchive = archive.Archive{
		Name:     RPMArchive.Name,
		Webpage:  RPMArchive.Webpage,
		Files:    RPMArchive.Files,
		Binaries: RPMArchive.Binaries,
		RuntimeDeps: []string{
			"/bin/sh",
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
		Conflicts:   RPMArchive.Conflicts,
		Description: RPMArchive.Description,
	}
)
