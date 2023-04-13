package archive

import (
	"packaging/pkg/build"

	"dagger.io/dagger"
)

type winArchive struct {
	a            NewArchive
	mirrorPrefix string
}

func NewWinArchive(a *NewArchive, mp string) Interface {
	if a == nil {
		panic("nil archive supplied")
	}

	return &winArchive{
		a:            *a,
		mirrorPrefix: mp,
	}
}

func (w *winArchive) Package(client *dagger.Client, c *dagger.Container, project *build.Spec) *dagger.Directory {
	dir := client.Directory()
	rootDir := "/package"

	c = c.WithDirectory(rootDir, dir)
	c = w.moveStaticFiles(c, rootDir)

	c = c.
		WithEnvVariable("PROJECT", project.Pkg).
		WithEnvVariable("VERSION", project.Tag).
		WithExec([]string{"bash", "-xuec", `
        : ${PROJECT}
        : ${VERSION}

        mkdir -p "/out"
        cd /package
        zip "/out/${PROJECT}-${VERSION}.zip" *
        `})

	return c.Directory("/out")
}

func (w *winArchive) moveStaticFiles(c *dagger.Container, rootdir string) *dagger.Container {
	for i := range w.a.Binaries[PkgKindWin] {
		b := w.a.Binaries[PkgKindWin][i]
		c = c.WithExec([]string{"cp", b, "/package"})
	}

	return c
}
