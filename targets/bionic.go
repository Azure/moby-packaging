package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	BionicRef            = path.Join(MirrorPrefix(), "buildpack-deps:bionic")
	BionicAptCacheKey    = "bionic-apt-cache"
	BionicAptLibCacheKey = "bionic-apt-lib-cache"
)

func Bionic(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	client = client.Pipeline("bionic/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(BionicRef)
	c = apt.Install(c, client.CacheVolume(BionicAptCacheKey), client.CacheVolume(BionicAptLibCacheKey), BaseBionicPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "bionic", pkgKind: "deb", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
