package targets

import (
	"context"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/tdnf"
)

const Mariner2Ref = "mcr.microsoft.com/cbl-mariner/base/core:2.0"

func Mariner2(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
	client = client.Pipeline("mariner2/" + string(platform))
	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Mariner2Ref)
	c = tdnf.Install(c, BaseMarinerPackages...)

	buildPlatform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return nil, err
	}

	attributes := StaticTargetAttributes["mariner2"]
	t := &Target{client: client, c: c, platform: platform, name: "mariner2", targetAttributes: attributes, buildPlatform: buildPlatform}
	t, err = t.WithPlatformEnvs().InstallGo(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}
