package targets

import (
	"context"
	"packaging/pkg/apt"
	"path"

	"dagger.io/dagger"
)

var (
	JammyRef            = path.Join(MirrorPrefix(), "buildpack-deps:jammy")
	JammyAptCacheKey    = "jammy-apt-cache"
	JammyAptLibCacheKey = "jammy-apt-lib-cache"
)

func Jammy(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("jammy/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(JammyRef)
	c = apt.AptInstall(c, client.CacheVolume(JammyAptCacheKey), client.CacheVolume(JammyAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "jammy", pkgKind: "deb", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
