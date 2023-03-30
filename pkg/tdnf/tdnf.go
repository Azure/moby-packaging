package tdnf

import "dagger.io/dagger"

func Install(c *dagger.Container, pkgs ...string) *dagger.Container {
	exec := []string{"tdnf", "install", "-y"}
	exec = append(exec, pkgs...)
	return c.WithExec(exec)
}
