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

func Bookworm(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("bookworm/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BookwormRef)
	c = apt.Install(c, client.CacheVolume(BookwormAptCacheKey), client.CacheVolume(BookwormAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	attributes := StaticTargetAttributes["bookworm"]
	t := &Target{client: client, c: c, platform: platform, name: "bookworm", targetAttributes: attributes, buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
