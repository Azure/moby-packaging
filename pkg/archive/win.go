package archive

import (
	"strings"

	"dagger.io/dagger"
)

type WinPackager struct {
	a            Archive
	mirrorPrefix string
}

func NewWinPackager(a *Archive, mp string) *WinPackager {
	if a == nil {
		panic("nil archive supplied")
	}

	return &WinPackager{
		a:            *a,
		mirrorPrefix: mp,
	}
}

func (w *WinPackager) Package(client *dagger.Client, c *dagger.Container, project *Spec) *dagger.Directory {
	dir := client.Directory()
	rootDir := "/package"
	sanitizedArch := strings.ReplaceAll(project.Arch, "/", "")

	c = c.WithDirectory(rootDir, dir)
	c = w.moveStaticFiles(c, rootDir)

	c = c.
		WithEnvVariable("PROJECT", project.Pkg).
		WithEnvVariable("VERSION", project.Tag).
		WithEnvVariable("REVISION", project.Revision).
		WithEnvVariable("ARCH", sanitizedArch).
		WithExec([]string{"bash", "-xuec", `
        : ${PROJECT}
        : ${VERSION}

        mkdir -p "/out"
        cd /package
        zip "/out/${PROJECT}-${VERSION}+azure-u${REVISION}.${ARCH}.zip" *
        `})

	return c.Directory("/out")
}

func (w *WinPackager) moveStaticFiles(c *dagger.Container, rootdir string) *dagger.Container {
	for i := range w.a.WinBinaries {
		b := w.a.WinBinaries[i]
		c = c.WithExec([]string{"cp", b, "/package"})
	}

	return c
}
