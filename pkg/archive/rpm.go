package archive

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"dagger.io/dagger"
)

var (
	nothing empty

	rpmDistroMap = map[string]string{
		"centos7":  "el7",
		"rhel8":    "el8",
		"rhel9":    "el9",
		"mariner2": "cm2",
	}

	rpmArchMap = map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}

	rpmPkgBlacklist = mapSet{
		"rhel9": {
			"libcgroup": nothing,
		},
	}
)

type (
	empty       struct{}
	mapSet      map[string]map[string]empty
	RpmPackager struct {
		a            Archive
		mirrorPrefix string
	}
)

func (m mapSet) contains(distro, pkg string) bool {
	if mm, ok := m[distro]; ok {
		if _, ok := mm[pkg]; ok {
			return true
		}
	}

	return false
}

func NewRPMPackager(a *Archive, mp string) *RpmPackager {
	if a == nil {
		panic("nil archive supplied")
	}

	return &RpmPackager{
		a:            *a,
		mirrorPrefix: mp,
	}
}

func (r *RpmPackager) Package(client *dagger.Client, c *dagger.Container, project *Spec) *dagger.Directory {
	dir := client.Directory()
	rootDir := "/package"

	c = c.WithDirectory(rootDir, dir)
	c = r.moveStaticFiles(c, rootDir)

	pkgDir := c.Directory(rootDir)
	fpm := fpmContainer(client, r.mirrorPrefix)

	filename := fmt.Sprintf("%s-%s+azure-%s.%s.%s.rpm", project.Pkg, project.Tag, project.Revision, rpmDistroMap[project.Distro], rpmArchMap[project.Arch])

	fpmArgs := []string{"fpm",
		"-s", "dir",
		"-t", "rpm",
		"-n", project.Pkg,
		"--version", project.Tag + "+azure",
		"--iteration", project.Revision,
		"--rpm-dist", rpmDistroMap[project.Distro],
		"--architecture", strings.Replace(project.Arch, "/", "", -1),
		"--description", r.a.Description,
		"--url", r.a.Webpage,
	}

	for i := range r.a.RuntimeDeps[PkgKindRPM] {
		dep := r.a.RuntimeDeps[PkgKindRPM][i]
		if rpmPkgBlacklist.contains(project.Distro, dep) {
			continue
		}

		fpmArgs = append(fpmArgs, "-d", dep)
	}

	for i := range r.a.Conflicts[PkgKindRPM] {
		conf := r.a.Conflicts[PkgKindRPM][i]
		fpmArgs = append(fpmArgs, "--conflicts", conf)
	}

	var args []string
	c, args = r.withInstallScripts(c)

	fpmArgs = append(fpmArgs, args...)
	fpmArgs = append(fpmArgs, ".")

	return fpm.WithDirectory("/package", pkgDir).
		WithDirectory("/build", c.Directory("/build")).
		WithWorkdir("/package").
		WithEnvVariable("OUTPUT_FILENAME", filename).
		WithExec(fpmArgs).
		WithExec([]string{"bash", "-c", `mkdir -vp /out; mv *.rpm "/out/${OUTPUT_FILENAME}"`}).
		Directory("/out")
}

func (r *RpmPackager) withInstallScripts(c *dagger.Container) (*dagger.Container, []string) {
	newArgs := []string{}

	for i := range r.a.InstallScripts[PkgKindRPM] {
		script := r.a.InstallScripts[PkgKindRPM][i]
		var a []string
		c, a = r.installScript(&script, c)
		newArgs = append(newArgs, a...)
	}

	return c, newArgs
}

func (r *RpmPackager) installScript(script *InstallScript, c *dagger.Container) (*dagger.Container, []string) {
	newArgs := []string{}

	var templateStr, filename, flag string
	switch script.When {
	case PkgActionPostInstall:
		filename = filenamePostInstall
		flag = flagPostInstall
		templateStr = `
{{ replace .Script "\n" "\n  " }}
            `
	case PkgActionUpgrade:
		filename = filenamePostUpgrade
		flag = flagUpgrade
		templateStr = `
if [ $1 -ge 1 ]; then
  {{ replace .Script "\n" "\n  " }}
fi
            `
	case PkgActionPreRemoval, PkgActionPostRemoval:
		filename = filenamePreRm
		flag = flagPreRm
		templateStr = `
if [ $1 -eq 0 ]; then
  {{ replace .Script "\n" "\n  " }}
fi
            `
	default:
		panic("unrecognized package action: " + fmt.Sprintf("%d", script.When))
	}

	filename = filepath.Join("/build", filename)

	tpl, err := template.New("installScript").Funcs(template.FuncMap{"replace": strings.ReplaceAll}).Parse(templateStr)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, script)
	if err != nil {
		panic(err)
	}

	c = c.WithNewFile(filename, dagger.ContainerWithNewFileOpts{Contents: buf.String()})
	newArgs = append(newArgs, flag, filename)
	return c, newArgs
}
func (r *RpmPackager) moveStaticFiles(c *dagger.Container, rootdir string) *dagger.Container {
	files := r.a.Files
	for i := range r.a.Systemd {
		sd := r.a.Systemd[i]
		files = append(files, File{
			Source: sd.Source,
			Dest:   sd.Dest,
		})
	}

	for i := range files {
		f := files[i]
		c = f.MoveStaticFile(c, rootdir)
	}

	return c
}
