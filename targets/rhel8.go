package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
)

var (
	Rhel8Ref = path.Join(MirrorPrefix(), "almalinux:8")
)

func Rhel8(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("rhel8/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Rhel8Ref).
		WithExec([]string{"bash", "-c", `
        yum -y install dnf-plugins-core
        yum config-manager --set-enabled powertools
        yum install -y gcc-toolset-12-binutils
        `})
	c = YumInstall(c, BaseRPMPackages...)
	c = c.WithEnvVariable("GCC_VERSION", "12").
		WithEnvVariable("GCC_ENV_VILE", "/opt/rh/gcc-toolset-12/enable")

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	attributes := StaticTargetAttributes["rhel8"]
	t := &Target{client: client, c: c, platform: platform, name: "rhel8", targetAttributes: attributes, buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
