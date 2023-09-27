package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	BookwormRef            = path.Join(MirrorPrefix(), "buildpack-deps:bookworm")
	BookwormAptCacheKey    = "bookworm-apt-cache"
	BookwormAptLibCacheKey = "bookworm-apt-lib-cache"
)

func Bookworm(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	client = client.Pipeline("bookworm/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BookwormRef)
	c = apt.Install(c, client.CacheVolume(BookwormAptCacheKey), client.CacheVolume(BookwormAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "bookworm", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
