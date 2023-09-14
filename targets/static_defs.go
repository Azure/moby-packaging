package targets

type TargetAttributes struct {
	PkgKind string

	// The file extension of a target. Will be one of ["deb", "rpm", "win"]
	Extension string

	// Used for generating a filename. For example, the bookworm distro would
	// have "debian" in the filename, where as the rhel9 distro would have
	// "el9"
	OsComponent string

	// Used for generating a filename. For example, the bookworm distro would
	// have "12" in the filename, jammy would have "22.04", and mariner2 would
	// have "cm2"
	VersionComponent string
}

var StaticTargetAttributes = map[string]TargetAttributes{
	"bookworm": {
		PkgKind:          "deb",
		Extension:        "deb",
		OsComponent:      "debian",
		VersionComponent: "12",
	},
	"bullseye": {
		PkgKind:          "deb",
		Extension:        "deb",
		OsComponent:      "debian",
		VersionComponent: "11",
	},
	"buster": {
		PkgKind:          "deb",
		Extension:        "deb",
		OsComponent:      "debian",
		VersionComponent: "10",
	},
	"focal": {
		PkgKind:          "deb",
		Extension:        "deb",
		OsComponent:      "ubuntu",
		VersionComponent: "20.04",
	},
	"jammy": {
		PkgKind:          "deb",
		Extension:        "deb",
		OsComponent:      "ubuntu",
		VersionComponent: "22.04",
	},
	"rhel9": {
		PkgKind:          "rpm",
		Extension:        "rpm",
		OsComponent:      "el9",
		VersionComponent: "el9",
	},
	"rhel8": {
		PkgKind:          "rpm",
		Extension:        "rpm",
		OsComponent:      "el8",
		VersionComponent: "el8",
	},
	"centos7": {
		PkgKind:          "rpm",
		Extension:        "rpm",
		OsComponent:      "el7",
		VersionComponent: "el7",
	},
	"mariner2": {
		PkgKind:          "rpm",
		Extension:        "rpm",
		OsComponent:      "cm2",
		VersionComponent: "cm2",
	},
	"windows": {
		PkgKind:     "win",
		Extension:   "zip",
		OsComponent: "windows",
		// VersionComponent: "", // Not used
	},
}
