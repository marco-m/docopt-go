# Changelog of the fork github.com/marco-m/docopt-go

## Release 0.8.0 UNRELEASED DATE

**NOTICE** This releases abandons backward compatibility with https://github.com/docopt/docopt.go

### New

- Function `Parse` that never calls `os.Exit()`.
- Function `MustParse`.
- Error `ErrHelp` (user requested help).
- Add CHANGELOG.

### Breaking changes

- Remove deprecated `Parse` function.
- Remove `ParseArgs` function; see new `Parse`.

### Changes

- Use Go 1.21.

## Release 0.7.0 2022-03-30

### New

- Add `go.mod`, use Go 1.18 and update import path to this fork.
- Build: add Taskfile support (like make, but simpler).
