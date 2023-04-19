package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/apt"
)

var (
	WindowsRef = path.Join(MirrorPrefix(), "buildpack-deps:bullseye")
)

func Windows(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("windows/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(WindowsRef)
	c = apt.AptInstall(c, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), BaseWinPackages...)
	c = c.WithEnvVariable("GOOS", "windows")

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "windows", pkgKind: "win", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
