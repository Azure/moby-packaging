package containerd

import (
	"os"

	"dagger.io/dagger"
)

var (
	Repo = "https://github.com/containerd/containerd.git"
	Ref  = "31aa4358a36870b21a992d3ad2bef29e1d693bec"
)

func FetchRef(client *dagger.Client) *dagger.GitRef {
	repo := envOrDefault("CONTAINERD_REPO", Repo)
	ref := envOrDefault("CONTAINERD_COMMIT", Ref)
	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref)
}

func envOrDefault(name string, def string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return def
}
