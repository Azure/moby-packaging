package containerd

import (
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
)

func GoVersion(spec *archive.Spec) string {
	return goversion.OneTwentyThree
}
