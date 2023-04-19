package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	BullseyeRef            = path.Join(MirrorPrefix(), "buildpack-deps:bullseye")
	BullseyeAptCacheKey    = "bullseye-apt-cache"
	BullseyeAptLibCacheKey = "bullseye-apt-lib-cache"
)

func Bullseye(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("bullseye/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BullseyeRef)
	c = apt.AptInstall(c, client.CacheVolume(BullseyeAptCacheKey), client.CacheVolume(BullseyeAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "bullseye", pkgKind: "deb", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
