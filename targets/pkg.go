package targets

import (
	"context"
	"strings"

	"dagger.io/dagger"
	"github.com/Azure/moby-packaging/pkg/archive"
)

const (
	hcsShimGitRepo = "https://github.com/Microsoft/hcsshim.git"
)

func FetchRef(client *dagger.Client, repo, commit string) *dagger.GitRef {
	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(commit)
}

func (t *Target) getSource(project *archive.Spec) *dagger.Directory {
	gitRef := FetchRef(t.client, project.Repo, project.Commit)
	dir := fetchExternalSource(t.client, gitRef, project)

	return dir
}

func fetchExternalSource(client *dagger.Client, gitRef *dagger.GitRef, project *archive.Spec) *dagger.Directory {
	switch project.Pkg {
	case "moby-containerd":
		if project.Distro == "windows" {
			return injectHCSShimSource(client, gitRef, project)
		}
	}

	return gitRef.Tree()
}

func injectHCSShimSource(client *dagger.Client, gitRef *dagger.GitRef, project *archive.Spec) *dagger.Directory {
	c := client.Container().
		From(MirrorPrefix()+"/buildpack-deps:buster").
		WithDirectory("/src", gitRef.Tree())

	commit, err := c.
		WithDirectory("/out", client.Directory()).
		WithWorkdir("/src").
		WithExec([]string{"awk", `/Microsoft\/hcsshim/{ print $2 >"/out/COMMIT" }`, "go.mod"}).
		File("/out/COMMIT").
		Contents(context.TODO())

	if err != nil {
		panic(err)
	}

	commit = strings.Trim(commit, " \n\t\r")

	hcsShimSourceDir := FetchRef(client, hcsShimGitRepo, commit).Tree()
	dir := c.WithDirectory("/src/hcs-shim", hcsShimSourceDir).Directory("/src")

	return dir
}
