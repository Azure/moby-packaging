package archive

import (
	"errors"
	"fmt"
)

// type PkgKind int
type PkgKind string
type PkgKindMap map[PkgKind][]string
type PkgAction int
type PkgInstallMap map[PkgKind][]InstallScript
type PkgInstallMap2 map[PkgKind][]InstallScript2

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

type InstallScript2 struct {
	When   PkgAction `json:"when"`
	Script Text      `json:"script"`
}

type Archive struct {
	Name    string `json:"name"`
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

type NewArchive struct {
	Name     Text      `json:"name"`
	Makefile Text      `json:"makefile"`
	Webpage  Text      `json:"webpage"`
	Files    []File    `json:"files"`
	Systemd  []Systemd `json:"systemd"`
	// required by some package types for dependency resolution
	Binaries       PkgKindMap     `json:"binaries"`
	Recommends     PkgKindMap     `json:"recommends"`
	Suggests       PkgKindMap     `json:"suggests"`
	Conflicts      PkgKindMap     `json:"conflicts"`
	Replaces       PkgKindMap     `json:"replaces"`
	Provides       PkgKindMap     `json:"provides"`
	BuildDeps      PkgKindMap     `json:"buildDeps"`
	RuntimeDeps    PkgKindMap     `json:"runtimeDeps"`
	InstallScripts PkgInstallMap2 `json:"installScripts"`
	Description    Text           `json:"description"`
}

var (
	ErrUnknownPkgAction = errors.New("unrecognized package type")
)

func (a *PkgAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	ret := PkgAction(-1)
	switch s {
	case "pre-removal":
		ret = PkgActionPreRemoval
	case "post-removal":
		ret = PkgActionPostRemoval
	case "post-install":
		ret = PkgActionPostInstall
	case "post-upgrade":
		ret = PkgActionUpgrade
	default:
		return fmt.Errorf("%w: %s", ErrUnknownPkgAction, s)
	}

	*a = ret
	return nil
}
