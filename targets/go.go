package targets

import (
	"context"
	"fmt"
	"path"

	"dagger.io/dagger"
)

var (
	GoVersion     = "1.19.5"
	GoRef         = path.Join("mcr.microsoft.com/oss/go/microsoft/golang:" + GoVersion)
	GoModCacheKey = "go-mod-cache"
)

func InstallGo(ctx context.Context, c *dagger.Container, modCache, buildCache *dagger.CacheVolume) (*dagger.Container, error) {
	dir := c.From(GoRef).Directory("/usr/local/go")
	pathEnv, err := c.EnvVariable(ctx, "PATH")
	if err != nil {
		return nil, fmt.Errorf("error getting PATH: %w", err)
	}
	if pathEnv == "" {
		return nil, fmt.Errorf("PATH is empty")
	}

	return c.WithDirectory("/usr/local/go", dir).
		WithEnvVariable("PATH", "/go/bin:/usr/local/go/bin:"+pathEnv).
		WithEnvVariable("GOROOT", "/usr/local/go").
		WithEnvVariable("GOPATH", "/go").
		WithMountedCache("/root/.cache/go-build", buildCache).
		WithMountedCache("/go/pkg/mod", modCache), nil
}

func (t *Target) InstallGo(ctx context.Context) (*Target, error) {
	c, err := InstallGo(ctx, t.c, t.client.CacheVolume(GoModCacheKey), t.client.CacheVolume(t.name+"-go-build-cache-"+string(t.platform)))
	if err != nil {
		return nil, err
	}
	return t.update(c), nil
}
