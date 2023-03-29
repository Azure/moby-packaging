package archive

import (
	"packaging/pkg/build"

	"dagger.io/dagger"
)

type Interface interface {
	Package(client *dagger.Client, c *dagger.Container, spec *build.Spec) *dagger.Directory
}
