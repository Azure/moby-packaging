package runc

import (
	"os"

	"dagger.io/dagger"
)

var (
	Repo = "https://github.com/opencontainers/runc.git"
	Ref  = "5fd4c4d144137e991c4acebb2146ab1483a97925"
)

func FetchRef(client *dagger.Client) *dagger.GitRef {
	repo := envOrDefault("RUNC_REPO", Repo)
	ref := envOrDefault("RUNC_COMMIT", Ref)
	return client.Git(repo, dagger.GitOpts{KeepGitDir: true}).Commit(ref)
}

func envOrDefault(name string, def string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	return def
}
