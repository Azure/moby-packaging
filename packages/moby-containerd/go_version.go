package containerd

import (
	"strings"

	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
	"github.com/Masterminds/semver/v3"
)

func GoVersion(spec *archive.Spec) string {
	version, _, _ := strings.Cut(spec.Tag, "~")
	v, err := semver.NewVersion(version)
	if err != nil {
		panic(err)
	}

	if v.Major() < 2 {
		return goversion.OneTwentyOne
	} else {
		return goversion.OneTwentyTwo
	}
}
