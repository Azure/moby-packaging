The builds are kicked off by JSON data in the form of an array of
`archive.Spec` structs. These are passed to the build as template parameters.
There is no way for ADO to validate the format of an `Object` template
parameter, so it has to be validated before it is used to ensure it won't cause
problems later on in the build.

In the future, we should provide more robust validation, possibly using cue.

```
Usage:
  -project string
    	name of project
  -revision string
    	revision for build set
  -tag string
    	tag for build set
```

Look at the code for the current validation rules. There is a good chance that this README is out of date, but at the time of its writing the rules were:

1. All specs must have the same package name
1. All specs must have the same tag
1. All specs must have the same revision number
1. All specs must have the same commit
1. All specs must have the same repo
1. The `Pkg` value for all specs must match the value passed in as the `--project` argument
1. The `Tag` value for all specs must match the value passed in as the `--tag` argument
1. The `Revision` value for all specs must match the value passed in as the `--revision` argument
1. No spec may have a blank value
1. No spec value may contain a single quote (`'`), as the spec is injected into
   bash scripts surrounded by single quotes.
