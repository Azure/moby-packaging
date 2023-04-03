package targets

import (
	"context"
	"os"
	buildx "packaging/moby-buildx"
	cli "packaging/moby-cli"
	compose "packaging/moby-compose"
	containerd "packaging/moby-containerd"
	shim "packaging/moby-containerd-shim-systemd"
	engine "packaging/moby-engine"
	mobyinit "packaging/moby-init"
	runc "packaging/moby-runc"
	"packaging/pkg/apt"
	"packaging/pkg/archive"
	"packaging/pkg/build"
	"strings"

	"dagger.io/dagger"
)

func (t *Target) AptInstall(pkgs ...string) *Target {
	c := apt.AptInstall(t.c, t.client.CacheVolume(t.name+"-apt-cache"), t.client.CacheVolume(t.name+"-apt-lib-cache"), pkgs...)
	return t.update(c)
}

type Target struct {
	c        *dagger.Container
	name     string
	platform dagger.Platform
	client   *dagger.Client
	pkgKind  string

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

func GetTarget(distro string) func(context.Context, *dagger.Client, dagger.Platform) (*Target, error) {
	switch distro {
	case "jammy":
		return Jammy
	case "buster":
		return Buster
	case "bionic":
		return Bionic
	case "bullseye":
		return Bullseye
	case "focal":
		return Focal
	case "rhel8":
		return Rhel8
	case "rhel9":
		return Rhel9
	case "centos7":
		return Centos7
	case "windows":
		return Windows
	case "mariner2":
		return Mariner2
	default:
		panic("unknown distro: " + distro)
	}
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

func (t *Target) goMD2Man() *dagger.File {
	repo := "https://github.com/cpuguy83/go-md2man.git"
	ref := "v2.0.2"
	outfile := "/build/bin/go-md2man"
	d := t.client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref).Tree()
	c := t.client.Container().From(GoRef).
		WithDirectory("/build", d).
		WithWorkdir("/build").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", outfile})

	return c.File(outfile)
}

func (t *Target) Packager(projectName string) archive.Interface {
	mappings := map[string]archive.Archive{
		"moby-engine":                  engine.Archive,
		"moby-cli":                     cli.Archive,
		"moby-containerd":              containerd.Archive,
		"moby-containerd-shim-systemd": shim.Archive,
		"moby-runc":                    runc.Archive,
		"moby-compose":                 compose.Archive,
		"moby-buildx":                  buildx.Archive,
		"moby-init":                    mobyinit.Archive,
	}

	a := mappings[projectName]

	switch t.PkgKind() {
	case "deb":
		return archive.NewDebArchive(&a, MirrorPrefix())
	case "rpm", "mariner2-rpm":
		return archive.NewRPMArchive(&a, MirrorPrefix())
	case "win":
		return archive.NewWinArchive(&a, MirrorPrefix())
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

func (t *Target) Make(project *build.Spec) *dagger.Directory {
	projectDir := t.client.Host().Directory(project.Pkg)
	hackDir := t.client.Host().Directory("hack/cross")
	md2man := t.goMD2Man()

	source := t.getSource(project)
	commitTime := t.getCommitTime(project.Pkg, source)

	build := t.c.Pipeline(project.Pkg).
		WithDirectory("/build", projectDir).
		WithDirectory("/build/debian/legal", projectDir.Directory("legal")).
		WithDirectory("/build/hack/cross", hackDir).
		WithDirectory("/build/src", source).
		WithWorkdir("/build").
		WithMountedFile("/usr/bin/go-md2man", md2man).
		WithEnvVariable("REVISION", project.Revision).
		WithEnvVariable("VERSION", project.Tag).
		WithEnvVariable("COMMIT", project.Commit).
		WithEnvVariable("SOURCE_DATE_EPOCH", commitTime).
		WithExec(t.applyPatchesCommand()).
		WithExec([]string{"/usr/bin/make", t.PkgKind()})
		// WithExec([]string{"mkdir", "/out"}).
		// WithExec([]string{"tar", "-cvzf", "/out/test.tar.gz", "/build"})

	//return build.Directory("/out")

	packager := t.Packager(project.Pkg)
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
