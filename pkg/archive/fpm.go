package archive

import (
	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

func fpmContainer(client *dagger.Client, mirrorPrefix string) *dagger.Container {
	c := client.Container().
		From(mirrorPrefix + "/debian:bullseye")
	c = apt.Install(c, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), "ruby", "build-essential", "rpm")
	return c.WithExec([]string{"gem", "install", "fpm"})
}
