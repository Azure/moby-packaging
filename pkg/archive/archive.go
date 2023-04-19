package archive

type PkgKind string
type PkgKindMap map[PkgKind][]string
type PkgAction int
type PkgInstallMap map[PkgKind][]InstallScript

const (
	PkgKindDeb PkgKind = "deb"
	PkgKindRPM PkgKind = "rpm"
	PkgKindWin PkgKind = "win"
)

const (
	flagPostInstall     = "--after-install"
	flagUpgrade         = "--after-upgrade"
	flagPreRm           = "--before-remove"
	flagPostRm          = "--after-remove"
	filenamePostInstall = "postinst"
	filenamePostUpgrade = "postup"
	filenamePreRm       = "prerm"
	filenamePostRm      = "postrm"
)

const (
	PkgActionPreRemoval PkgAction = iota
	PkgActionPostRemoval
	PkgActionPostInstall
	PkgActionUpgrade
)

type InstallScript struct {
	When   PkgAction
	Script string
}

type Archive struct {
	Name    string
	Webpage string
	Files   []File
	Systemd []Systemd
	// list of filenames
	Postinst []string
	// required for debian dependency resolution
	Binaries       []string
	WinBinaries    []string
	Recommends     PkgKindMap
	Suggests       PkgKindMap
	Conflicts      PkgKindMap
	Replaces       PkgKindMap
	Provides       PkgKindMap
	BuildDeps      PkgKindMap
	RuntimeDeps    PkgKindMap
	InstallScripts PkgInstallMap
	Description    string
}
