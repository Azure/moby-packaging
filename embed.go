package main

import (
	"embed"
	"io/fs"
	"path/filepath"

	"dagger.io/dagger"
)

var (
	//go:embed hack/cross
	hackCrossFS embed.FS

	//go:embed packages
	packagesFS embed.FS
)

func packageDir(client *dagger.Client, name string) *dagger.Directory {
	root := client.Directory()

	dir, err := fs.Sub(packagesFS, filepath.Join("packages", name))
	if err != nil {
		panic(err)
	}

	err = fs.WalkDir(dir, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			root = root.WithNewDirectory(path, dagger.DirectoryWithNewDirectoryOpts{Permissions: int(info.Mode().Perm())})
			return nil
		}

		dt, err := fs.ReadFile(dir, path)
		if err != nil {
			return err
		}
		root = root.WithNewFile(path, string(dt), dagger.DirectoryWithNewFileOpts{Permissions: int(info.Mode().Perm())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return root
}

func hackCrossDir(client *dagger.Client) *dagger.Directory {
	root := client.Directory()

	dir, err := fs.Sub(hackCrossFS, "hack/cross")
	if err != nil {
		panic(err)
	}

	err = fs.WalkDir(dir, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			root = root.WithNewDirectory(path, dagger.DirectoryWithNewDirectoryOpts{Permissions: int(info.Mode().Perm())})
			return nil
		}

		dt, err := fs.ReadFile(dir, path)
		if err != nil {
			return err
		}
		root = root.WithNewFile(path, string(dt), dagger.DirectoryWithNewFileOpts{Permissions: int(info.Mode().Perm())})
		return nil
	})
	if err != nil {
		panic(err)
	}
	return root
}
