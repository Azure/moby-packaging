package targets

import (
	"context"
	"os"
	"packaging/pkg/apt"

	"dagger.io/dagger"
)

var (
	Repo = "https://github.com/krallin/tini.git"
	Ref  = "de40ad007797e0dcd8b7126f27bb87401d224240"
)

func envOrDefault(name string, def string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return def
}

func fetchRef(client *dagger.Client) *dagger.GitRef {
	repo := envOrDefault("TINI_REPO", Repo)
	ref := envOrDefault("TINI_COMMIT", Ref)
	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref)
}

func fpmContainer(client *dagger.Client) *dagger.Container {
	c := client.Container().
		From(MirrorPrefix() + "/debian:bullseye")
	c = apt.AptInstall(c, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), "ruby", "build-essential")
	return c.WithExec([]string{"gem", "install", "fpm"})
}

func MakeInit(ctx context.Context, client *dagger.Client, arch, distro string) error {
	// fetch source
	srcDir := fetchRef(client).Tree()

	// build tini for arch

	build := client.Container().From(MirrorPrefix() + "/debian:bullseye")
	build = apt.AptInstall(build, client.CacheVolume("bullseye-apt-cache"), client.CacheVolume("bullseye-apt-lib-cache"), "cmake", "vim-common")

	build = build.WithDirectory("/tini", srcDir).
		WithWorkdir("/tini/build").
		WithExec([]string{"bash", "-c", `
        cmake ..
        make tini-static
        `})

	tini := build.File("/tini/build/tini-static")

	// arch may need translator
	// package according to distro
	_, err := fpmContainer(client).
		WithMountedFile("/workspace/docker-init", tini).
		WithWorkdir("/workspace").
		WithExec([]string{"fpm",
			"-s", "dir",
			"-t", "deb",
			"-n", "moby-init",
			"--license", "none",
			"--version", "0.0.1",
			"--architecture", arch,
			"--description", "test description",
			"--url", "https://example.com",
			"docker-init=/usr/bin/docker-init",
		}).
		Directory("/workspace").
		Export(ctx, "./out")
	return err
}
