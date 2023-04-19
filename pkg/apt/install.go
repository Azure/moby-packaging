package apt

import (
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

const (
	aptUpdatedPath = "/var/cache/apt/moby/updated"
)

func Install(c *dagger.Container, aptCache, aptLibCache *dagger.CacheVolume, pkgs ...string) *dagger.Container {
	// We don't want these files to persist in the rootfs, so we create them in a tempdir and mount them in.
	dir := c.Directory("/etc/apt/apt.conf.d").
		WithNewFile("docker-clean", "").        // overwrite the default docker-clean file
		WithNewFile("docker-gzip-indexes", ""). // overwrite the default docker-gzip-indexes file
		WithNewFile("keep-cache", "Binary::apt::APT::Keep-Downloaded-Packages \"true\";").
		WithNewFile("update-success", fmt.Sprintf(`APT::Update::Post-Invoke-Success { "mkdir -p %s; touch %s"; };`, filepath.Dir(aptUpdatedPath), aptUpdatedPath))

	c2 := c.WithMountedDirectory("/etc/apt/apt.conf.d", dir)

	if aptCache != nil {
		c2 = c2.WithMountedCache("/var/cache/apt", aptCache, dagger.ContainerWithMountedCacheOpts{
			Sharing: dagger.Locked,
		})

	}
	if aptLibCache != nil {
		c2 = c2.WithMountedCache("/var/lib/apt", aptLibCache, dagger.ContainerWithMountedCacheOpts{
			Sharing: dagger.Locked,
		})
	}

	c2 = c2.WithExec([]string{
		"/bin/sh", "-ec", `
			export DEBIAN_FRONTEND=noninteractive
			UPDATED_PATH=` + aptUpdatedPath + `
			UPDATED_MAX_AGE=60
			if [ -z "$(find ${UPDATED_PATH} -mmin -${UPDATED_MAX_AGE})" ]; then
				apt-get update
			fi
			apt-get install -y ` + strings.Join(pkgs, " "),
	})

	return c.WithRootfs(c2.Rootfs())
}
