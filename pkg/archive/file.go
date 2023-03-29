package archive

import (
	"path/filepath"

	"dagger.io/dagger"
)

type File struct {
	Source   string
	Dest     string
	IsDir    bool
	Compress bool
}

func (f *File) MoveStaticFile(c *dagger.Container, rootdir string) *dagger.Container {
	dest := filepath.Join(rootdir, f.Dest)

	if f.IsDir && f.Source == "" {
		return c.WithExec([]string{"mkdir", "-p", dest})
	}

	destDir := filepath.Dir(dest)

	// Sometimes required for manpages nested within directories
	// this may not be necessary
	if f.Compress && f.IsDir {
		return c.
			WithEnvVariable("SOURCE", f.Source).
			WithEnvVariable("DEST", dest).
			WithExec([]string{"bash", "-exuc", `
            : ${SOURCE}
            : ${DEST}

            if [ -L "$SOURCE" ]; then
                SOURCE="$(readlink "$SOURCE")"
            fi

            mkdir -p "$DEST"

            export SOURCE DEST
            find "$SOURCE" -type f -printf "%P\0" | xargs -0I{} bash -c '
                    [ -z "{}" ] && exit 0

                    prefix="$(dirname "{}")"
                    mkdir -p "$DEST/$prefix"
                    gzip -c "$SOURCE/{}" > "$DEST/{}.gz"
                '
            `,
			})
	}

	if f.Compress {
		return c.
			WithEnvVariable("SOURCE", f.Source).
			WithEnvVariable("DEST_DIR", destDir).
			WithEnvVariable("DEST", dest).
			WithExec([]string{"bash", "-exuc", `
            : ${SOURCE}
            : ${DEST_DIR}
            : ${DEST}

            if [ -L "$SOURCE" ]; then
                SOURCE="$(readlink "$SOURCE")"
            fi

            install -d "$DEST_DIR"
            gzip -c "$SOURCE" > "$DEST"
            `,
			})
	}

	return c.
		WithEnvVariable("SOURCE", f.Source).
		WithEnvVariable("DEST_DIR", destDir).
		WithEnvVariable("DEST", dest).
		WithExec([]string{"bash", "-exuc", `
        : ${SOURCE}
        : ${DEST_DIR}
        : ${DEST}

        install -d "$DEST_DIR"
        cp -Lr "$SOURCE" "$DEST"
        `,
		})
}
