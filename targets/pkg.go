package targets

import (
	"packaging/pkg/build"

	"dagger.io/dagger"
)

const (
	hcsShimGitRepo = "https://github.com/Microsoft/hcsshim.git"
)

func FetchRef(client *dagger.Client, repo, commit string) *dagger.GitRef {
	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(commit)
}

func (t *Target) getSource(project *build.Spec) *dagger.Directory {
	client := t.client.Pipeline(project.Pkg + "-src")
	gitRef := FetchRef(client, project.Repo, project.Commit)

	return gitRef.Tree()
}
