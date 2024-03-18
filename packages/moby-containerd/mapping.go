package containerd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Masterminds/semver/v3"
)

var (
	//go:embed postinstall/deb/postinstall
	debPostInstall string
	//go:embed postinstall/deb/prerm
	debPreRm string
	//go:embed postinstall/deb/postrm
	debPostRm string

	//go:embed postinstall/rpm/postinstall
	rpmPostInstall string
	//go:embed postinstall/rpm/prerm
	rpmPreRm string
	//go:embed postinstall/rpm/upgrade
	rpmUpgrade string
)

func Archives(version string) (map[string]archive.Archive, error) {
	// We use `~` in packaging to indicate that the version is a pre-release version.
	// semver does not recognize `~`.
	// We only really care about major/minor version here, so we can just cut off the pre-release part.
	version, _, _ = strings.Cut(version, "~")
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s: %w", version, err)
	}
	switch fmt.Sprintf("%d.%d", v.Major(), v.Minor()) {
	case "1.6", "1.7":
		return Archives_1_X, nil
	case "2.0":
		return Archives_2_0, nil
	default:
		return nil, fmt.Errorf("unsupported version: %s", version)
	}
}
