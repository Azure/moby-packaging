package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
)

var (
	Rhel9Ref = path.Join(MirrorPrefix(), "almalinux:9")
)

func Rhel9(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("rhel9/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Rhel9Ref).
		WithExec([]string{"bash", "-c", `
        yum -y install dnf-plugins-core || true
        yum config-manager --set-enabled powertools || true
        yum config-manager --enable powertools || yum config-manager --enable resilientstorage
        yum config-manager --enable crb || true
        `})
	c = YumInstall(c, BaseRPMPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "rhel9", pkgKind: "rpm", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil

}
