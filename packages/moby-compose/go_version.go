package compose

import (
	"strings"

	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
	"github.com/Masterminds/semver/v3"
)

const (
	threshold = "2.22.0"
)

func GoVersion(s *archive.Spec) string {
	tag, _, _ := strings.Cut(s.Tag, "~")

	v, err := semver.NewVersion(tag)
	if err != nil {
		return goversion.DefaultVersion
	}

	t, err := semver.NewVersion(threshold)
	if err != nil {
		return goversion.DefaultVersion
	}

	if v.Compare(t) >= 0 { // if v >= t
		return goversion.OneTwentyThree
	}

	return goversion.DefaultVersion
}
