package archive

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"dagger.io/dagger"
)

func join(a []string) string {
	return strings.Join(a, ", ")
}

const ControlTemplate = `
Source: {{ .Name }}
Section: admin
Priority: optional
Maintainer: Microsoft <support@microsoft.com>
Build-Depends: bash-completion,
               go-md2man <!cross>,
               go-md2man:amd64 <cross>,
               pkg-config, {{ join .BuildDeps }}
Rules-Requires-Root: no
Homepage: {{ .Webpage }}

Package: {{ .Name }}
Architecture: linux-any
Depends: ${misc:Depends}, ${shlibs:Depends}, {{ join .RuntimeDeps }}
Recommends: {{ join .Recommends }}
Conflicts: {{ join .Conflicts }}
Replaces: {{ join .Replaces }}
Provides: {{ join .Provides }}
Description: {{ .Description }}
`

var (
	DebDistroMap = map[string]string{
		"xenial":  "ubuntu16.04",
		"yakkety": "ubuntu16.10",
		"zesty":   "ubuntu17.04",
		"artful":  "ubuntu17.10",
		"bionic":  "ubuntu18.04",
		"cosmic":  "ubuntu18.10",
		"disco":   "ubuntu19.04",
		"eoan":    "ubuntu19.10",
		"focal":   "ubuntu20.04",
		"groovy":  "ubuntu20.10",
		"hirsute": "ubuntu21.04",
		"impish":  "ubuntu21.10",
		"jammy":   "ubuntu22.04",
		"kinetic": "ubuntu22.10",
		"noble":   "ubuntu24.04",
		"lunar":   "ubuntu23.04",

		"buster":   "debian10",
		"bullseye": "debian11",
		"bookworm": "debian12",
		"trixie":   "debian13",
		"forky":    "debian14",
	}
)

type DebPackager struct {
	a            Archive
	mirrorPrefix string
}

func NewDebPackager(a *Archive, mp string) *DebPackager {
	if a == nil {
		panic("nil archive supplied")
	}

	return &DebPackager{
		a:            *a,
		mirrorPrefix: mp,
	}
}

func (d *DebPackager) Package(client *dagger.Client, c *dagger.Container, project *Spec) *dagger.Directory {
	dir := client.Directory()
	rootDir := "/package"

	version := fmt.Sprintf("%s-%su%s", project.Tag, DebDistroMap[project.Distro], project.Revision)
	c = c.WithDirectory(rootDir, dir)
	c = d.moveStaticFiles(c, rootDir)
	c = d.withControlFile(c, version, project)

	pkgDir := c.Directory(rootDir)

	fpmArgs := []string{"fpm",
		"-s", "dir",
		"-t", "deb",
		"-n", project.Pkg,
		"--version", version,
		"--architecture", strings.Replace(project.Arch, "/", "", -1),
		"--deb-custom-control", "/build/control",
	}

	var newArgs []string
	c, newArgs = d.withInstallScripts(c)

	fpmArgs = append(fpmArgs, d.systemdArgs()...)
	fpmArgs = append(fpmArgs, newArgs...)
	fpmArgs = append(fpmArgs, ".")

	fpm := fpmContainer(client, d.mirrorPrefix)
	return fpm.WithDirectory("/package", pkgDir).
		WithDirectory("/build", c.Directory("/build")).
		WithWorkdir("/package").
		WithExec(fpmArgs).
		WithExec([]string{"bash", "-ec", `mkdir -vp /out; mv *.deb /out`}).
		Directory("/out")
}

func (d *DebPackager) moveStaticFiles(c *dagger.Container, rootdir string) *dagger.Container {
	for i := range d.a.Files {
		f := d.a.Files[i]
		c = f.MoveStaticFile(c, rootdir)
	}

	return c
}

func (d *DebPackager) withInstallScripts(c *dagger.Container) (*dagger.Container, []string) {
	newArgs := []string{}

	for i := range d.a.InstallScripts {
		script := d.a.InstallScripts[i]
		var a []string
		c, a = d.installScript(&script, c)
		newArgs = append(newArgs, a...)
	}

	return c, newArgs
}

func (d *DebPackager) installScript(script *InstallScript, c *dagger.Container) (*dagger.Container, []string) {
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
{{ replace .Script "\n" "\n  " }}
            `
	case PkgActionPreRemoval:
		filename = filenamePreRm
		flag = flagPreRm
		templateStr = `
{{ replace .Script "\n" "\n  " }}
            `
	case PkgActionPostRemoval:
		filename = filenamePostRm
		flag = flagPostRm
		templateStr = `
{{ replace .Script "\n" "\n  " }}
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

	c = c.WithNewFile(filename, buf.String())
	newArgs = append(newArgs, flag, filename)
	return c, newArgs
}

func (d *DebPackager) withControlFile(c *dagger.Container, version string, project *Spec) *dagger.Container {
	t := ControlTemplate

	tpl, err := template.New("control").Funcs(template.FuncMap{"join": join}).Parse(t)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	err = tpl.Execute(buf, d.a)
	if err != nil {
		panic(err)
	}

	return c.
		WithNewFile("/build/debian/control", buf.String()).
		WithEnvVariable("PROJECT_NAME", project.Pkg).
		WithEnvVariable("VERSION", version).
		WithEnvVariable("DISTRO", project.Distro).
		WithEnvVariable("_BINARIES", strings.Join(d.a.Binaries, " ")).
		WithExec([]string{
			"bash", "-exuc", `
        : ${PROJECT_NAME}
        : ${VERSION}
        : ${DISTRO}
        : ${_BINARIES}
        cat /build/debian/control

        BINARIES=($_BINARIES)

        cat <<EOF > debian/changelog
$PROJECT_NAME ($VERSION) $DISTRO; urgency=low
  * Version: 1.0
 -- Microsoft <support@microsoft.com>  Mon, 12 Mar 2018 00:00:00 +0000
EOF

        args=()
        for b in "${BINARIES[@]}"; do
            args+=(-e "$b")
        done

        dpkg-shlibdeps "${args[@]}"
        dpkg-gencontrol -P/package -Ocontrol
        `,
		})
}

func (d *DebPackager) systemdArgs() []string {
	args := []string{}

	for i := range d.a.Systemd {
		sd := d.a.Systemd[i]
		args = append(args, "--deb-systemd", sd.Source)
	}

	// this could change
	args = append(args,
		"--deb-systemd-enable",
		"--deb-systemd-auto-start",
		"--deb-systemd-restart-after-upgrade",
	)

	return args
}
