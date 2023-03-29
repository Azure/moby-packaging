package targets

import (
	"context"
	"packaging/pkg/apt"
	"path"

	"dagger.io/dagger"
)

var (
	BusterRef            = path.Join(MirrorPrefix(), "buildpack-deps:buster")
	BusterAptCacheKey    = "buster-apt-cache"
	BusterAptLibCacheKey = "buster-apt-lib-cache"
)

func Buster(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("buster/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BusterRef)
	c = apt.AptInstall(c, client.CacheVolume(BusterAptCacheKey), client.CacheVolume(BusterAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "buster", pkgKind: "deb", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
