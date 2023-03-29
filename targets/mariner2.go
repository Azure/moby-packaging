package targets

const Mariner2Ref = "mcr.microsoft.com/cbl-mariner/base/core:2.0"

// package targets

// import (
// 	"context"
// 	"path"

// 	"dagger.io/dagger"
// )

// var (
// 	Rhel8Ref = path.Join(MirrorPrefix(), "almalinux:8")
// )

// func Rhel8(ctx context.Context, client *dagger.Client, platform dagger.Platform) (*Target, error) {
// 	client = client.Pipeline("rhel8/" + string(platform))
// 	c := client.Container(dagger.ContainerOpts{Platform: platform}).From(Rhel8Ref).
// 		WithExec([]string{"bash", "-c", `
//         yum -y install dnf-plugins-core || true
//         yum config-manager --set-enabled powertools || true
//         `})
// 	c = YumInstall(c, BaseRPMPackages...)

// 	buildPlatform, err := client.DefaultPlatform(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	t := &Target{client: client, c: c, platform: platform, name: "rhel8", pkgKind: "rpm", buildPlatform: buildPlatform}
// 	t, err = t.WithPlatformEnvs().InstallGo(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return t, nil
// }
