package archive

import (
	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/build"
)

type Interface interface {
	Package(client *dagger.Client, c *dagger.Container, spec *build.Spec) *dagger.Directory
}
