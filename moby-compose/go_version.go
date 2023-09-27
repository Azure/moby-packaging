package compose

import (
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
	"github.com/Masterminds/semver/v3"
)

const (
	threshold = "2.22.0"
)

func GoVersion(s *archive.Spec) string {
	v, err := semver.NewVersion(s.Tag)
	if err != nil {
		return goversion.DefaultVersion
	}

	t, err := semver.NewVersion(threshold)
	if err != nil {
		return goversion.DefaultVersion
	}

	if v.Compare(t) >= 0 { // if v >= t
		return goversion.OneTwentyOne
	}

	return goversion.DefaultVersion
}
