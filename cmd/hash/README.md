This utility produces a consistent filename from three pieces of information.
This program is intended to produce filenames for files that represent
information about a group of packages. Its intended first use-case is to
provide a consistent filename for the file containing a JSON list of
`archive.Spec` structs. The current setup is to group the building of packages
by:

1. Project name
1. Tag (release version)
1. Revision

For all builds containing the same project name AND the same tag AND the same
revision number, the input will be an array of `archive.Specs`. All specs in
that array MUST have the same project name, tag, and revision. Because these
are all the same, and they must be unique, those three pieces of information
can uniquely represent the build group. As such, they are the inputs to this
program and determine its output.

Consistent filenames are important because they avoid relying too heavily on
ADO templating which is difficult to read and reason about. Because the specs
remain immutable throughout the build process, we can calculate values
deterministically throughout the build to consistently reference files named by
this program.

```
Usage:
  -project string
    	name of the project
  -tag string
    	tag of artifact
  -revision string
    	revision
```
