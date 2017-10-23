# Symbols Extractor Design & Hacking Guide

This guide will teach you how *Symbols Extractor* is designed and implemented
and how you can extend its functionality.

## Contents

## Directory structure

*Symbols Extractor* is situated in the following directories; each directory is
provided with a short description of its meaning:
```
<your-golang-workspace>/
    src/
        github.com/gofed/symbols-extractor/
            .git/
            cgo/
            cmd/
            docs/
            pkg/
                parser/
                    alloctable/
                        global/
                    expression/
                    file/
                    statement/
                    symboltable/
                        global/
                        stack/
                    testdata/
                    type/
                    types/
                testing/
                    utils/
                types/
            symboltables/
            tests/
                bin/
                pkg/
                src/
                    specs/
                        templates/
            vendor/
```

In what follow next, we use the following names for crucial directories:
* under the `<project-root>`, we mean
  `<your-golang-workspace>/src/github.com/gofed/symbols-extractor`
* under the `<tests-root>`, we mean `<project-root>/tests`
* under the `<tests-specs>`, we mean `<tests-root>/src/specs`

## Testing

### Test specification format

All the specification of tests are distributed over various YAML files stored
in directory `<tests-specs>`. In `<tests-specs>` should be placed a YAML file
that will be recognized as the entry point for a test generator. Its default
name is `tests.yml` and it could contains following entries:
```yaml
spec-version:
    # Mandatory; specifies the version of the test specification format.
info:
    # Mandatory; provides informations about test set.
config:
    # Optional; here are options for a test generator.
groups:
    # Optional; definitions of all test groups together with their subgroups.
    # If there are no groups, all tests are considered under the main (root)
    # group.
tests:
    # Mandatory; test set itself.
```
Moreover, if there is also a file named `all.yml` in the same directory as
`tests.yml`, a test generator reads `all.yml` instead. Inside `all.yml` is a
list of `tests.yml`-like files that are processed by a test generator in the
order of their appearance in the list. In `all.yml` file, if an item is an
existing file or directory, it is processed (in the case of directory, every
file in the directory is processed in the alphabetical order but one exception
--- if there is `all.yml` file, the whole scenario is recursively repeated). If
such item does not exist and its name is suffixless, the suffix `.yml` is added
and the item is tried to be reopened. If it fails again, error should be
reported. A test generator utility provides a various flags that can customize
so far described behavior.

`spec-version`'s value is a sequence of 2 or 3 numbers delimited by dots with
the meaning as major level, minor level, and optionally patch level,
respectively.
```
Version     = MajorLever "." MinorLevel [ "." PatchLevel ] .
MajorLevel  = VersionPart .
MinorLevel  = VersionPart .
PatchLevel  = VersionPart .

VersionPart = ( "0" ... "9" ) { "0" ... "9" } .
```

`info` is a mapping containing following keys:
```yaml
name:
    # Mandatory; the name of the test set.
author:
    # Mandatory; a list of authors.
version:
    # Mandatory; the version of the test set.
comment:
    # Optional; additional comment on the test set.
```

`config` is a mapping with the following entries:
```yaml
commands:
    # Optional; path to commands directory, relatively to <tests-specs>.
    # Default value is "commands".
templates:
    # Optional; path to the directory with templates, relatively to
    # <tests-specs>. Default value is "templates".
```

`groups` is a list of mappings where every mapping contains:
```yaml
name:
    # Mandatory; a name of group. Must be unique in the current set.
subgroups:
    # Optional; a list of subgroups. Each subgroup must be defined in the
    # moment of reaching the end of `groups` scope.
```

Finally, `tests` contains a list of tests. As a test is considered a mapping
containing following keys:
```yaml
name:
    # Mandatory; a name of test.
group:
    # Optional; a group name to which the test is belonging to. Missing or
    # empty group means the test belongs to the main (root) group.
comment:
    # Optional; this test's commentary.
fileset:
    # Mandatory; a list of definitions of files that are used as test's input
    # data.
testers:
    # Mandatory; a list of testers that do all the test-work on the in this
    # test specified file set, gather results and write them to final report.
```

A file definition inside a `fileset` is a mapping containing the following
keys:
```yaml
name:
    # Mandatory; a path to file.
version:
    # Mandatory; a file/package/project version.
content:
    # Mandatory; a file content.
```
A real path of the file is computed from `<tests-root>`, test set `name`,
groups, test `name`, and `name` and `version` fields. If a chain of groups to
which the test belongs to is `A <- B <- C`, where `A <- B` means that `B` is a
subgroup of `A`, then the real path to test file will be:
```
<test-root>/src/<test set name>/A/B/C/<test name>/v<version>/<name>
```

A tester is a mapping with the following entries:
```yaml
command:
    # Mandatory; a command that runs the test.
expected-result:
    # Mandatory; an expected result to which the result of the finished test
    # run will be compared.
```
Both `command` and `expected-result` fields are scalars written in the specific
format that will be explained later.

### Generators

To reduce the writing same or very similar things again, generators can be
used. A generator is a mapping with only one key:
```yaml
generate:
    # Mandatory; commands that generate output from the given input.
```
Commands in `generate` field have the same format as those from tester's
`command` and `expected-result` fields. Going from the top to the bottom,
generators can be used in those places:
* in `tests` field in a test set entry-point file; here, a generator is capable
  of to create and add new tests to the test set;
* in `fileset` and `testers` fields in a test;
* in a `content` field in a file definition;
* in an `expected-result` field in a tester.

A generator is invoked once it is find and when its parameters and environment
is known.

### Commands and templates

Both command and template are mappings with the following keys:
```yaml
name:
    # Mandatory; a command or template name.
definition:
    # Mandatory; a command or template definition.
```

A command is distinguished from a template by its placement. Commands are
stored in files in `commands` directory (or in which user specify) and
similarly templates are stored in files in `templates` directory.

While a template definition is a scalar that must met the requirements for
[Go templates](https://golang.org/pkg/text/template/), the definition of
command is written in a simplified version of Go language.
