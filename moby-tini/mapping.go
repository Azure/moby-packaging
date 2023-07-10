package mobyinit

import "github.com/Azure/moby-packaging/pkg/archive"

var (
	Archives = map[string]archive.Archive{
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
		Name:    "moby-tini",
		Webpage: "https://github.com/krallin/tini",
		Files: []archive.File{
			{
				Source: "/build/src/build/tini-static",
				Dest:   "usr/bin/docker-init",
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
		Conflicts: []string{
			"tini",
		},
		Replaces: []string{
			"tini",
		},
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
	}

	MarinerArchive = RPMArchive
)
