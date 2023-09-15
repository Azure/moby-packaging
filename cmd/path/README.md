This utility handles filename and directory-name generation for packages.
Historically, packages were built and stored at a specific path which encoded
information about the artifact. We are moving towards a system where the
information about an artifact is stored in the spec file, and everything is
generated from that same source.

When given a spec file, this utility will generate the `basename` of the
package (example `moby-containerd_1.7.0-ubuntu22.04u7_amd64.deb`), the `dir`
where the package will be stored (example `<root>/jammy/linux_amd64`), or both
(the `full-path`).

```
Usage: go run ./cmd/path [basename|dir|full-path] --spec-file=SPEC_FILE [--bundle-dir=BUNDLE_DIR]
  -bundle-dir string (OPTIONAL)
    	base directory of bundled files
  -spec-file string (REQUIRED)
    	path of spec file
```

All subcommands require the `--spec-file` argument. The `--bundle-dir` argument
is optional; in addition, it has no effect on the `basename` subcommand. If it
is not provided, the `dir` and `full-path` subcommands will produce a relative
path *without* the leading `./`.
