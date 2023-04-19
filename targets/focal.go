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

func Focal(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("focal/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(FocalRef)
	c = apt.AptInstall(c, client.CacheVolume(FocalAptCacheKey), client.CacheVolume(FocalAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "focal", pkgKind: "deb", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
