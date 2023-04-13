## About

This repository holds the logic for building and packaging various moby OSS
packages, for several different distros. Builds are performed by fetching the
source from upstream (github) and building it according to the logic specified
in the Makefiles. Once built, a project is packaged for its target distribution.

This project uses dagger to manage containerized building and packaging.

## Quick start

The following example shows how to create a .deb package for containerd v1.7.0,
specifically for ubuntu jammy.

```bash
cat > ./moby-containerd.json <<'EOF'
{
  "arch": "amd64",
  "commit": "1fbd70374134b891f97ce19c70b6e50c7b9f4e0d",
  "repo": "https://github.com/containerd/containerd.git",
  "package": "moby-containerd",
  "distro": "jammy",
  "tag": "1.7.0",
  "os": "linux",
  "revision": "7"
}
EOF

go run packaging --build-spec=./moby-containerd.json
```

The `commit` field is the commit hash of the `tag` in question. The `tag` field
is used when deriving the filename and linker flags, but the `commit` is the
source of truth for the source code to be built.

This will create a file,
`bundles/jammy/moby-containerd_1.7.0+azure-ubuntu22.04u7_amd64.deb`, which can
then be published in a package repository.


## Adding new packages

### Overview

Currently, adding a new package involves several steps. We plan to improve the
user experience around this in the near future.

### Add a new package directory

In the root of this repository, create a new directory for the project you want
to build.

```bash
mkdir -p moby-init
```

#### Container filesystem layout

moby-packaging will create and manage a pipeline of containers with an
opinionated filesystem layout. Within the container, anything in the project
directory (`moby-init` in this example) will be mounted into the `/build`
directory within the container.

The source for the target package (in this example, `moby-init` which is built
from the upstream repository [krallin/tini](https://github.com/krallin/tini))
will be mounted into `/build/src`.

This will be important later, when [specifying](#specify-the-package-layout) the
layout of the package.

Any static files that you want to be in the final package should live in
the target package's directory (again, `moby-init` in this example).

### Add Makefiles to the package directory

Capture the build logic in a Makefile. You will need a Make target for each
package type you wish to build (currently, `deb`, `rpm`, or `win`).

```bash
cat > moby-init/Makefile <<'EOF'

.PHONY: rpm deb

rpm deb: tini-static

tini-static:
	mkdir -vp ./src/build
	cd ./src/build && \
		cmake .. && \
		make tini-static
EOF
```

Note that the working directory will be `/build` (see [Container filesystem
layout](#container-filesystem-layout)). The above reference to `./src/build`
is at the absolute path `/build/src/build`.

This particular build will output a file, `tini-static` in the absolute path
`/build/src/build`.

### Specify the package layout

Packages for distro repositories are essentially archives (tarballs, cpio, zip
files) containing files at their final destination on the target system. They additionally
contain additional information, such as post-install scripts (to be run on the target system),
dependency information, and a description.

To specify where our newly built binary should go, we have to tell
moby-packaging where to find them in our container, and where they belong on
the target system. Here is an example for our (admittedly simple) package.

```bash
cat > moby-init/mapping.go <<'EOF'
package mobyinit

import "packaging/pkg/archive"

var (
	Archive = archive.Archive{
		Name:    "moby-init",
		Webpage: "https://github.com/krallin/tini",
		Files: []archive.File{
			{
				Source: "/build/src/build/tini-static",
				Dest:   "/usr/bin/docker-init",
			},
		},
		Binaries: []string{
			"/build/src/build/tini-static",
		},
		Conflicts: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"tini",
			},
		},
		Replaces: archive.PkgKindMap{
			archive.PkgKindDeb: {
				"tini",
			},
		},
		Description: `tiny but valid init for containers
 Tini is the simplest init you could think of.
 .
 All Tini does is spawn a single child (Tini is meant to be run in a
 container), and wait for it to exit all the while reaping zombies and
 performing signal forwarding.`,
	}
)
EOF
```

The struct here is defined in `pkg/archive/archive.go`.

The key element here is the `Files` entry: the `Source` file is the location in
the build container of a file we want to package. The `Dest` file is the final
location on the target system. Once built and published to a debian repo, one
would run `apt-get install moby-init`; this would install the `tini-static`
binary we built at the location `/ur/bin/docker-init`.

The `Conflicts` and `Replaces` entries are used by the consuming package manager
to remove older versions of the same package.

In addition to these two entries, there are entries which specify runtime
dependency packages. The package manager will install those packages as well.
The `Binaries` entry is also used for dependency management. Since a binary may
be dynamically linked, it will be inspected for runtime dependencies (and also
installed by the package manager).

`Name`, `Webpage`, and `Description` are used by the package manager when
displaying information about the package.

Finally, note the `package` directive at the top of the file. `mobyinit` will
be used as an import in the next step.

### Updating moby-packaging to recognize the new package

To enable the packaging system to build this package, update
`targets/target.go`:

```go
import (
    // ...
    mobyinit "packaging/moby-init"
    // ...
)

// ...

func (t *Target) Packager(projectName string) archive.Interface {
	mappings := map[string]archive.Archive{
		"moby-engine":                  engine.Archive,
		"moby-cli":                     cli.Archive,
		"moby-containerd":              containerd.Archive,
		"moby-containerd-shim-systemd": shim.Archive,
		"moby-runc":                    runc.Archive,
		"moby-compose":                 compose.Archive,
		"moby-buildx":                  buildx.Archive,

         // this references the `Archive` struct created in the previous step
		"moby-init":                    mobyinit.Archive,
	}

	a := mappings[projectName]

	switch t.PkgKind() {
	case "deb":
		return archive.NewDebArchive(&a, MirrorPrefix())
	case "rpm":
		return archive.NewRPMArchive(&a, MirrorPrefix())
	case "win":
		return archive.NewWinArchive(&a, MirrorPrefix())
	default:
		panic("unknown pkgKind: " + t.pkgKind)
	}
}
```

### Producing the final package

As with the [quick start](#quick-start), we need to supply moby-packaging with
some information about what to build.

```bash
cat > ./moby-init.json <<'EOF'
{
  "arch": "amd64",
  "commit": "de40ad007797e0dcd8b7126f27bb87401d224240",
  "repo": "https://github.com/krallin/tini.git",
  "package": "moby-init",
  "distro": "jammy",
  "tag": "0.19.0",
  "os": "linux",
  "revision": "9"
}
EOF

go run packaging --build-spec=./moby-init.json
```

This will produce a package under `bundles/jammy` which is ready to deploy.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
