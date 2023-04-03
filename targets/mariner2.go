package targets

import (
	"context"
	"packaging/pkg/tdnf"

	"dagger.io/dagger"
)

const Mariner2Ref = "mcr.microsoft.com/cbl-mariner/base/core:2.0"

func Mariner2(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("mariner2/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Mariner2Ref).
		WithExec([]string{"bash", "-c", `
        yum -y install dnf-plugins-core || true
        `})
	c = tdnf.Install(c, BaseMarinerPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	t := &Target{client: client, c: c, platform: platform, name: "mariner2", pkgKind: "rpm", buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
