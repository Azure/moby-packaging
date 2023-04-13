package mobyinit

import "packaging/pkg/archive"

var (
	Mapping2 = []archive.File{
		{
			Source: "/build/src/build/tini-static",
			Dest:   "usr/bin/docker-init",
		},
		{
			Source: "/build/legal/LICENSE",
			Dest:   "/usr/share/doc/moby-init/LICENSE",
		},
		{
			Source:   "/build/legal/NOTICE",
			Dest:     "/usr/share/doc/moby-init/NOTICE.gz",
			Compress: true,
		},
	}

	Archive = archive.Archive{
		Name:    "moby-init",
		Webpage: "https://github.com/krallin/tini",
		Files: []archive.File{
			{
				Source: "/build/src/build/tini-static",
				Dest:   "usr/bin/docker-init",
			},
			{
				Source: "/build/legal/LICENSE",
				Dest:   "/usr/share/doc/moby-init/LICENSE",
			},
			{
				Source:   "/build/legal/NOTICE",
				Dest:     "/usr/share/doc/moby-init/NOTICE.gz",
				Compress: true,
			},
		},
		Binaries: []string{
			"/build/src/build/tini-static",
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"tini",
			},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"tini",
			},
		},
		Description: `tiny but valid init for containers
 Tini is the simplest init you could think of.
 .
 All Tini does is spawn a single child (Tini is meant to be run in a
 container), and wait for it to exit all the while reaping zombies and
 performing signal forwarding.`,
	}

	Dirs = []string{}
)
