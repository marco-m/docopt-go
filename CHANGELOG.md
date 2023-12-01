# Changelog of the fork github.com/marco-m/docopt-go

## Release 0.8.0 UNRELEASED DATE

**NOTICE** This releases abandons backward compatibility with https://github.com/docopt/docopt.go

### New

- Add function `Parse`, main entry point of the module. It never calls `os.Exit`.
- Add function `MustParse`, the only function that calls `os.Exit`.
- Add error `ErrHelp` (user requested help).
- Add CHANGELOG.
- Introduce `testscript` [1] to easily test executables, their output and their status code.

[1]: https://github.com/rogpeppe/go-internal/testscript

### Breaking changes

- Remove deprecated `Parse` function.
- Remove `ParseArgs` function; replaced by the new `Parse`.
- Remove field `HelpHandler` from struct `Parser`.
- Remove functions for `HelpHandler`: `PrintHelpAndExit`, `PrintHelpOnly`, `NoHelpHandler`.

### Changes

- Use Go 1.21.
- examples/unit_test: more idiomatic and simple Go unit tests.

## Release 0.7.0 2022-03-30

### New

- Add `go.mod`, use Go 1.18 and update import path to this fork.
- Build: add Taskfile support (like make, but simpler).
