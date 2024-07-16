package tini

import "github.com/Azure/moby-packaging/pkg/archive"

var (
	Archives = map[string]archive.Archive{
		"bookworm": DebArchive,
		"buster":   DebArchive,
		"bullseye": DebArchive,
		"bionic":   DebArchive,
		"focal":    DebArchive,
		"rhel8":    RPMArchive,
		"rhel9":    RPMArchive,
		"windows":  BaseArchive,
		"jammy":    DebArchive,
		"mariner2": MarinerArchive,
		"noble":    DebArchive,
	}

	BaseArchive = archive.Archive{
		Name:    "moby-tini",
		Webpage: "https://github.com/krallin/tini",
		Files: []archive.File{
			{
				Source: "/build/src/build/tini-static",
				Dest:   "usr/libexec/docker/docker-init",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-tini/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-tini/NOTICE.gz",
				Compress: true,
			},
		},
		Binaries: []string{
			"/build/src/build/tini-static",
		},
		Description: `tiny but valid init for containers
 Tini is the simplest init you could think of.
 .
 All Tini does is spawn a single child (Tini is meant to be run in a
 container), and wait for it to exit all the while reaping zombies and
 performing signal forwarding.`,
	}

	DebArchive = archive.Archive{
		Name:    BaseArchive.Name,
		Webpage: BaseArchive.Webpage,
		Files:   BaseArchive.Files,
		Binaries: []string{
			"/build/src/build/tini-static",
		},
		Conflicts:   []string{},
		Description: BaseArchive.Description,
	}

	RPMArchive = archive.Archive{
		Name:    BaseArchive.Name,
		Webpage: BaseArchive.Webpage,
		Files:   BaseArchive.Files,
		Binaries: []string{
			"/build/src/build/tini-static",
		},
		Description: BaseArchive.Description,
		Conflicts:   []string{},
	}

	MarinerArchive = RPMArchive
)
