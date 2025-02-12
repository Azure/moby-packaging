package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	BusterRef            = path.Join(MirrorPrefix(), "buildpack-deps:buster")
	BusterAptCacheKey    = "buster-apt-cache"
	BusterAptLibCacheKey = "buster-apt-lib-cache"
)

func Buster(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BusterRef)
	c = apt.Install(c, client.CacheVolume(BusterAptCacheKey), client.CacheVolume(BusterAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "buster", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
