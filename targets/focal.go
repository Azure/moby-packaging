package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	FocalRef            = path.Join(MirrorPrefix(), "buildpack-deps:focal")
	FocalAptCacheKey    = "focal-apt-cache"
	FocalAptLibCacheKey = "focal-apt-lib-cache"
)

func Focal(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(FocalRef)
	c = apt.Install(c, client.CacheVolume(FocalAptCacheKey), client.CacheVolume(FocalAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "focal", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}
	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
