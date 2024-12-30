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

func Bullseye(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BullseyeRef)
	c = apt.Install(c, client.CacheVolume(BullseyeAptCacheKey), client.CacheVolume(BullseyeAptLibCacheKey), BaseDebPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "bullseye", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
