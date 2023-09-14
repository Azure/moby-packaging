package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/Azure/moby-packaging/pkg/archive"
)

type allArgs struct {
	pkg      string
	tag      string
	revision string
}

func main() {
	args := allArgs{}
	flag.StringVar(&args.pkg, "project", "", "name of project")
	flag.StringVar(&args.tag, "tag", "", "tag for build set")
	flag.StringVar(&args.revision, "revision", "", "revision for build set")
	flag.Parse()

	if args.pkg == "" || args.tag == "" || args.revision == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := validate(args); err != nil {
		fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]%s\n", err)
		os.Exit(1)
	}
}

func validate(args allArgs) error {
	r := bufio.NewReader(os.Stdin)

	j, err := r.ReadBytes(byte(0))
	if err != nil && err != io.EOF {
		return fmt.Errorf("error: %s\n", err)
	}

	specs := []archive.Spec{}

	if err := json.Unmarshal(j, &specs); err != nil {
		return err
	}

	if len(specs) == 0 {
		return nil
	}

	lastCommit := specs[0].Commit
	lastRepo := specs[0].Repo
	for i := range specs {
		if project := specs[i].Pkg; project != args.pkg {
			return fmt.Errorf("package name does not match: was '%s', should be '%s'", args.pkg, project)
		}

		if tag := specs[i].Tag; tag != args.tag {
			return fmt.Errorf("package tag does not match: was '%s', should be '%s'", tag, args.tag)
		}

		if revision := specs[i].Revision; revision != args.revision {
			return fmt.Errorf("package revision does not match: was '%s', should be '%s'", revision, args.revision)
		}

		if commit := specs[i].Commit; commit != lastCommit {
			return fmt.Errorf("all builds should have the same commit hash: '%s' vs '%s'", commit, lastCommit)
		}

		if repo := specs[i].Repo; repo != lastRepo {
			return fmt.Errorf("all builds should have the same repo: '%s' vs '%s'", repo, lastCommit)
		}

		lastCommit = specs[i].Commit
		lastRepo = specs[i].Repo

		v := reflect.ValueOf(&specs[i]).Elem()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i).Interface().(string)

			if f == "" {
				return fmt.Errorf("blank value: %s", v.Type().Field(i).Name)
			}

			if strings.Contains(f, "'") {
				return fmt.Errorf(`illegal character in field '%s': "%s"`, v.Type().Field(i).Name, f)
			}
		}
	}

	return nil
}
