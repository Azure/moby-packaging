package engine

import (
	"github.com/Azure/moby-packaging/pkg/archive"
	"github.com/Azure/moby-packaging/pkg/goversion"
)

func GoVersion(_ *archive.Spec) string {
	return goversion.DefaultVersion
}
