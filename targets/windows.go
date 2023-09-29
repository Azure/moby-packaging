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

func Windows(ctx context.Context, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	client = client.Pipeline("windows/" + string(platform))

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	c := client.Container(dagger.ContainerOpts{Platform: buildPlatform}).From(WindowsRef)
	c = apt.Install(c, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), BaseWinPackages...)
	c = c.WithEnvVariable("GOOS", "windows")

	t := &Target{client: client, c: c, platform: platform, name: "windows", pkgKind: "win", buildPlatform: buildPlatform, goVersion: goVersion}

	t, err = t.WithPlatformEnvs().InstallGo(ctx, goVersion)
	if err != nil {
		return nil, err
	}

	return t, nil
}
