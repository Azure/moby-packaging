package targets

import (
	"context"
	"path"

	"dagger.io/dagger"
)

var (
	Centos7Ref = path.Join(MirrorPrefix(), "centos:7")
)

func Centos7(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("centos7/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Centos7Ref)
	c = YumInstall(c, BaseRPMPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	attributes := StaticTargetAttributes["centos7"]
	t := &Target{client: client, c: c, platform: platform, name: "centos7", targetAttributes: attributes, buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
