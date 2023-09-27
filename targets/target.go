package targets

import (
	"context"
	"fmt"
	"os"
	"strings"

	buildx "github.com/Azure/moby-packaging/moby-buildx"
	cli "github.com/Azure/moby-packaging/moby-cli"
	compose "github.com/Azure/moby-packaging/moby-compose"
	containerd "github.com/Azure/moby-packaging/moby-containerd"
	shim "github.com/Azure/moby-packaging/moby-containerd-shim-systemd"
	engine "github.com/Azure/moby-packaging/moby-engine"
	runc "github.com/Azure/moby-packaging/moby-runc"
	tini "github.com/Azure/moby-packaging/moby-tini"
	"github.com/Azure/moby-packaging/pkg/apt"
	"github.com/Azure/moby-packaging/pkg/archive"

	"dagger.io/dagger"
)

type GoVersionFunc = func(*archive.Spec) string

func (t *Target) AptInstall(pkgs ...string) *Target {
	c := apt.Install(t.c, t.client.CacheVolume(t.name+"-apt-cache"), t.client.CacheVolume(t.name+"-apt-lib-cache"), pkgs...)
	return t.update(c)
}

type Target struct {
	c         *dagger.Container
	name      string
	platform  dagger.Platform
	client    *dagger.Client
	pkgKind   string
	goVersion string

	buildPlatform dagger.Platform
}

func (t *Target) update(c *dagger.Container) *Target {
	return &Target{c: c, name: t.name, platform: t.platform, client: t.client, pkgKind: t.pkgKind}
}

func MirrorPrefix() string {
	prefix, ok := os.LookupEnv("MIRROR_PREFIX")
	if !ok {
		prefix = "mcr.microsoft.com/mirror/docker/library"
	}
	return prefix
}

type MakeTargetFunc func(context.Context, *dagger.Client, dagger.Platform, string) (*Target, error)

var targets = map[string]MakeTargetFunc{
	"jammy":    Jammy,
	"buster":   Buster,
	"bionic":   Bionic,
	"bullseye": Bullseye,
	"bookworm": Bookworm,
	"focal":    Focal,
	"rhel8":    Rhel8,
	"rhel9":    Rhel9,
	"centos7":  Centos7,
	"windows":  Windows,
	"mariner2": Mariner2,
}

func GetTarget(ctx context.Context, distro string, client *dagger.Client, platform dagger.Platform, goVersion string) (*Target, error) {
	f, ok := targets[distro]
	if !ok {
		panic("unknown distro: " + distro)
	}
	return f(ctx, client, platform, goVersion)
}

var (
	BaseWinPackages = []string{
		"binutils-mingw-w64",
		"g++-mingw-w64-x86-64",
		"gcc",
		"git",
		"make",
		"pkg-config",
		"quilt",
		"zip",
	}

	BaseDebPackages = []string{
		"build-essential",
		"cmake",
		"dh-make",
		"devscripts",
		"dh-apparmor",
		"dpkg-dev",
		"equivs",
		"fakeroot",
		"libbtrfs-dev",
		"libdevmapper-dev",
		"libltdl-dev",
		"libseccomp-dev",
		"quilt",
	}

	BaseBionicPackages = []string{
		"bash",
		"build-essential",
		"cmake",
		"dh-make",
		"devscripts",
		"dh-apparmor",
		"dpkg-dev",
		"equivs",
		"fakeroot",
		"libdevmapper-dev",
		"libltdl-dev",
		"libseccomp-dev",
		"quilt",
	}

	BaseMarinerPackages = []string{
		"bash",
		"binutils",
		"build-essential",
		"ca-certificates",
		"cmake",
		"device-mapper-devel",
		"diffutils",
		"dnf-utils",
		"file",
		"gcc",
		"git",
		"glibc-static",
		"libffi-devel",
		"libseccomp-devel",
		"libtool",
		"libtool-ltdl-devel",
		"make",
		"patch",
		"pkgconfig",
		"pkgconfig(systemd)",
		"rpm-build",
		"rpmdevtools",
		"selinux-policy-devel",
		"systemd-devel",
		"tar",
		"which",
		"yum-utils",
	}

	BaseRPMPackages = []string{
		"bash",
		"ca-certificates",
		"cmake",
		"device-mapper-devel",
		"gcc",
		"git",
		"glibc-static",
		"libseccomp-devel",
		"libtool",
		"libtool-ltdl-devel",
		"make",
		"make",
		"patch",
		"pkgconfig",
		"pkgconfig(systemd)",
		"rpmdevtools",
		"selinux-policy-devel",
		"systemd-devel",
		"tar",
		"which",
		"yum-utils",
	}

	GetGoVersionForPackage = map[string]GoVersionFunc{
		"moby-buildx":                  buildx.GoVersion,
		"moby-cli":                     cli.GoVersion,
		"moby-compose":                 compose.GoVersion,
		"moby-containerd":              containerd.GoVersion,
		"moby-containerd-shim-systemd": shim.GoVersion,
		"moby-engine":                  engine.GoVersion,
		"moby-runc":                    runc.GoVersion,
		"moby-tini":                    tini.GoVersion,
	}
)

func (t *Target) Container() *dagger.Container {
	return t.c
}

func (t *Target) WithExec(args []string, opts ...dagger.ContainerWithExecOpts) *Target {
	return t.update(t.c.WithExec(args, opts...))
}

func (t *Target) PkgKind() string {
	return t.pkgKind
}

func (t *Target) applyPatchesCommand() []string {
	return []string{
		"bash", "-exc", `
        [ -f patches/series ] || exit 0
        readarray -t patches < patches/series
        cd src/
        for f in "${patches[@]}"; do
            patch -p1 < "/build/patches/$f"
        done
        `,
	}
}

// Winres is used during windows builds (as part of the project build scripts) to "manifest" binaries.
// This is required for windows to properly identify the binaries.
func (t *Target) Winres() *dagger.File {
	return t.client.Container().
		From(GoRef).
		WithEnvVariable("GOBIN", "/build").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GO111MODULE", "on").
		WithExec([]string{"go", "install", "github.com/tc-hib/go-winres@v0.3.0"}).
		File("/build/go-winres")
}

func (t *Target) goMD2Man() *dagger.File {
	repo := "https://github.com/cpuguy83/go-md2man.git"
	ref := "v2.0.2"
	outfile := "/build/bin/go-md2man"
	srcDir := t.client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref).Tree()
	goRef := fmt.Sprintf("%s:%s", GoRepo, t.goVersion)

	c := t.client.Container().
		From(goRef).
		WithDirectory("/build", srcDir).
		WithWorkdir("/build").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", outfile})

	return c.File(outfile)
}

type Packager interface {
	Package(*dagger.Client, *dagger.Container, *archive.Spec) *dagger.Directory
}

func (t *Target) Packager(projectName, distro string) Packager {
	mappings := map[string]map[string]archive.Archive{
		"moby-engine":                  engine.Archives,
		"moby-cli":                     cli.Archives,
		"moby-containerd":              containerd.Archives,
		"moby-containerd-shim-systemd": shim.Archives,
		"moby-runc":                    runc.Archives,
		"moby-compose":                 compose.Archives,
		"moby-buildx":                  buildx.Archives,
		"moby-tini":                    tini.Archives,
	}

	as := mappings[projectName]
	a, ok := as[distro]
	if !ok {
		panic("unknown distro: " + distro)
	}

	switch t.PkgKind() {
	case "deb":
		return archive.NewDebPackager(&a, MirrorPrefix())
	case "rpm":
		return archive.NewRPMPackager(&a, MirrorPrefix())
	case "win":
		return archive.NewWinPackager(&a, MirrorPrefix())
	default:
		panic("unknown pkgKind: " + t.pkgKind)
	}
}

func (t *Target) getCommitTime(projectName string, sourceDir *dagger.Directory) string {
	commitTime, err := t.c.Pipeline(projectName+"/commit-time").
		WithMountedDirectory("/build/src", sourceDir).
		WithWorkdir("/build/src").
		WithExec([]string{"bash", "-ec", `date -u --date=$(git show -s --format=%cI HEAD) +%s > /tmp/COMMITTIME`}).
		File("/tmp/COMMITTIME").
		Contents(context.TODO())

	if err != nil {
		return ""
	}

	return strings.TrimSpace(commitTime)
}

func (t *Target) Make(project *archive.Spec) *dagger.Directory {
	projectDir := t.client.Host().Directory(project.Pkg)
	hackDir := t.client.Host().Directory("hack/cross")
	md2man := t.goMD2Man()

	source := t.getSource(project)
	commitTime := t.getCommitTime(project.Pkg, source)

	build := t.c.Pipeline(project.Pkg).
		WithDirectory("/build", projectDir).
		WithDirectory("/build/hack/cross", hackDir).
		WithDirectory("/build/src", source).
		WithWorkdir("/build").
		WithMountedFile("/usr/bin/go-md2man", md2man).
		WithMountedFile("/usr/bin/go-winres", t.Winres()).
		WithEnvVariable("TARGET_DISTRO", project.Distro).
		WithEnvVariable("REVISION", project.Revision).
		WithEnvVariable("VERSION", project.Tag).
		WithEnvVariable("COMMIT", project.Commit).
		WithEnvVariable("SOURCE_DATE_EPOCH", commitTime).
		WithExec(t.applyPatchesCommand()).
		WithExec([]string{"/usr/bin/make", t.PkgKind()})
		// WithExec([]string{"mkdir", "/out"}).
		// WithExec([]string{"tar", "-cvzf", "/out/test.tar.gz", "/build"})

	//return build.Directory("/out")

	packager := t.Packager(project.Pkg, project.Distro)
	return packager.Package(t.client, build, project)
}

func WithPlatformEnvs(c *dagger.Container, build, target dagger.Platform) *dagger.Container {
	split := strings.SplitN(string(build), "/", 2)
	buildOS := split[0]
	buildArch := split[1]
	var buildVarient string
	if len(split) == 3 {
		buildVarient = split[2]
	}

	split = strings.SplitN(string(target), "/", 2)
	targetOS := split[0]
	targetArch := split[1]
	var targetVariant string
	if len(split) == 3 {
		targetVariant = split[2]
	}

	return c.
		WithEnvVariable("BUILDARCH", buildArch).
		WithEnvVariable("BUILDVARIANT", buildVarient).
		WithEnvVariable("BUILDOS", buildOS).
		WithEnvVariable("BUILDPLATFORM", string(build)).
		WithEnvVariable("TARGETARCH", targetArch).
		WithEnvVariable("TARGETVARIANT", targetVariant).
		WithEnvVariable("TARGETOS", targetOS).
		WithEnvVariable("TARGETPLATFORM", string(target))
}

func (t *Target) WithPlatformEnvs() *Target {
	return t.update(WithPlatformEnvs(t.c, t.buildPlatform, t.platform))
}
