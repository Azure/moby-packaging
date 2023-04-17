## About

This project is an all-in-one system for building upstream projects from
source, and creating packages for various platforms and distributions. For
example, say you want to build and package `containerd` for use on an Ubuntu
system. Fill out a couple of YAML files, and moby-packaging will fetch the
source from the upstream git repository, build the project, and package it
into a `.deb` file for distribution.

Because such workflows are often automated, we have made it easy to trigger
builds externally.

<!--
This repository holds the logic for building and packaging various moby OSS
packages, for several different distros. Builds are performed by fetching the
source from upstream (github) and building it according to the logic specified
in the Makefiles. Once built, a project is packaged for its target distribution.
-->

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
  "revision": "1"
}
EOF

# moby-containerd/package.yml has already been provided in this repo
go run packaging \
    --build-spec=./moby-containerd.json \
    --package-spec=moby-containerd/package.yml \
    --project-dir=moby-containerd
```

The `commit` field is the commit hash of the `tag` in question. The `tag` field
is used when deriving the filename and linker flags, but the `commit` is the
source of truth for the source code to be built.

This will create a file,
`bundles/jammy/moby-containerd_1.7.0+azure-ubuntu22.04u7_amd64.deb`, which can
then be published in a package repository.


## Adding new packages

### Overview

A project is made up of three main components:

1. A project directory containing static files to be used during building and
   packaging.
1. A build spec, which provides information about the source to be built, the
   target OS and architecture, and versioning information.
1. A package definition file, in YAML format. This file specifies build steps
   (in the form of a Makefile), and information needed by the packager. For
   example, conflicting packages to remove before installation, or runtime
   dependencies that should be installed alongside the new package.

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

This particular build will output a file, `tini-static` at the absolute path
`/build/src/build/tini-static`.

### Specify the package layout

Packages for distro repositories are essentially archives (tarballs, cpio, zip
files) containing files at their final destination on the target system. They
also contain additional information, such as post-install scripts (to be run on
the target system), dependency information, and a description.

To specify where our newly built binary should go, we have to tell
moby-packaging where to find them in our container, and where they belong on
the target system. Here is an example for our (admittedly simple) package.

```bash
printf '
# the `name` field is for documentation only, and ignored. instead, the
# `package` field in the build spec is used for generating control information
name: moby-init
webpage: https://github.com/krallin/tini
makefile: "#moby-init/Makefile"
files:
  - source: /build/src/build/tini-static
    dest: usr/bin/docker-init
  - source: /build/legal/LICENSE
    dest: /usr/share/doc/moby-init/LICENSE
  - source: /build/legal/NOTICE
    dest: /usr/share/doc/moby-init/NOTICE.gz
    compress: true
binaries:
  deb:
    - /build/src/build/tini-static
  rpm:
    - /build/src/build/tini-static
  win: []
conflicts:
  deb:
    - tini
replaces:
  deb:
    - tini
description: |-
  tiny but valid init for containers
   Tini is the simplest init you could think of.
   .
   All Tini does is spawn a single child (Tini is meant to be run in a
   container), and wait for it to exit all the while reaping zombies and
   performing signal forwarding.
' > moby-init/package.yml
```

The format here is unmarshaled into the struct in `pkg/archive/archive.go`.

Strings that begin with `#` are *embedded files*. In this example,
`#moby-init/Makefile` means "replace this string with the contents of
moby-init/Makefile on the host filesystem". Thus, the whole string
"#moby-init/Makefile" will be replaced with the contents of the Makefile we
created in the previous step.

The key element here is the `files` entry: the `source` file is the location in
the build container of a file we want to package. the `dest` file is the final
location on the target system. Once built and published to a debian repo, one
would run `apt-get install moby-init`; this would install the built
`tini-static` binary at the location `/ur/bin/docker-init`.

the `conflicts` and `replaces` entries are used by the consuming package manager
to remove older versions of the same package.

In addition to these two entries, there are entries which specify runtime
dependency packages. The package manager will install those packages as well.
the `binaries` entry is also used for dependency management. Since a binary may
be dynamically linked, it will be inspected for runtime dependencies (and also
installed by the package manager).

`name`, `webpage`, and `description` are used by the package manager when
displaying information about the package.

See the [yaml schema](#yaml-input) for a more detailed description of the
various options.

### Producing the final package

As with the [quick start](#quick-start), we need to supply moby-packaging with
some information about the source and verioning info.

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

go run packaging \
    --build-spec=./moby-init.json \
    --package-spec=moby-init/package.yml \
    --project-dir=moby-init
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

## YAML input

| Key  | Value |
|------|-------|
| makefile | The text content or anchored filename of a Makefile. See Makefile formatting for more information |
| webpage | https://github.com/moby/moby |
| files | *[objects]*, where each dictionary has the following keys: |
|| **source**: *string* that represents the source path of the file to be installed. |
|| **dest**: *string* that represents the destination path of the file to be installed. |
|| **isdir** *(optional)*: *bool* that indicates whether the destination path is a directory. Default value is false. |
|| **compress** *(optional)*: *bool* that indicates whether the file should be compressed. Default value is false. |
|| *note*: when both `compress` and `isdir` are true, the individual files in the directory will be compressed, not the directory as a whole |
| systemd | *[objects]*, where each dictionary has the same keys as files, and represents the systemd units to be installed. |
| binaries | A list of binaries according to package type. This is used for dependency management, since the binary may be dynamically linked: |
|| **deb**: *[strings]*, where each string represents the path of a binary file to be installed for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the path of a binary file to be installed for Red Hat-based systems. |
|| **win**: *[strings]*, where each string represents the path of a binary file to be installed for Windows systems. |
| recommends | A list of recommended packages according to package type |
|| **deb**: *[strings]*, where each string represents the name of a recommended package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name of a recommended package for Red Hat-based systems. |
| suggests | A list of suggested packages according to package type: |
|| **deb**: *[strings]*, where each string represents the name of a suggested package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name of a suggested package for Red Hat-based systems. |
| conflicts | A list of conflicting packages according to package type: |
|| **deb**: *[strings]*, where each string represents the name of a conflicting package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name of a conflicting package for Red Hat-based systems. |
| replaces | list of packages that will be replaced by this one, according to package type: |
|| **deb**: *[strings]*, where each string represents the name of a replaced package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name of a replaced package for Red Hat-based systems. |
| provides | The names of packages provided by this package: |
|| **deb**: *[strings]*, where each string represents the name of the provided package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name of the provided package for Red Hat-based systems. |
| runtimeDeps | A list of runtime dependencies, according to package type: |
|| **deb**: *[strings]*, where each string represents the name and version of a runtime dependency package for Debian-based systems. |
|| **rpm**: *[strings]*, where each string represents the name and version of a runtime dependency package for Red Hat-based systems. |
| installScripts | A list of install scripts, according to package type and when they should be executed: |
|| **deb**: *[objects]*, where each dictionary has the following keys: |
|| **when**: *string* that represents when the script should be executed (e.g. post-install, pre-removal, post-removal). |
|| **script**: *string* that contains the path to the script to be executed. |
|| **rpm**: *[objects]* |

