package targets

import "dagger.io/dagger"

func YumInstall(c *dagger.Container, pkgs ...string) *dagger.Container {
	exec := []string{"yum", "install", "--skip-broken", "-y"}
	exec = append(exec, pkgs...)
	return c.WithExec(exec)
}
