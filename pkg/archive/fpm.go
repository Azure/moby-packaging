package archive

import (
	"packaging/pkg/apt"

	"dagger.io/dagger"
)

func fpmContainer(client *dagger.Client, mirrorPrefix string) *dagger.Container {
	c := client.Container().
		From(mirrorPrefix + "/debian:bullseye")
	c = apt.AptInstall(c, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), "ruby", "build-essential", "rpm")
	return c.WithExec([]string{"gem", "install", "fpm"})
}
