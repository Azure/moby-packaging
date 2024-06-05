package containerd

import (
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
	"github.com/Masterminds/semver/v3"
)

func GoVersion(spec *archive.Spec) string {
	v, err := semver.NewVersion(spec.Tag)
	if err != nil {
		panic(err)
	}

	if v.Major() < 2 {
		return goversion.OneTwentyOne
	} else {
		return goversion.OneTwentyTwo
	}
}
