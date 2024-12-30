package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	NobleRef            = path.Join(MirrorPrefix(), "buildpack-deps:noble")
	NobleAptCacheKey    = "noble-apt-cache"
	NobleAptLibCacheKey = "noble-apt-lib-cache"
)

func Noble(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(NobleRef)
	c = apt.Install(c, client.CacheVolume(NobleAptCacheKey), client.CacheVolume(NobleAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "noble", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
