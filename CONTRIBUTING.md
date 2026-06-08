# Contributing

RefactorLah is a conservative refactoring tool. Contributions should preserve the core rule: rewrite only references that can be proven from syntax and project configuration.

## Tests

Format Go source files:

```bash
bin/format.sh
```

Run the full test suite:

```bash
bin/test.sh
```

`bin/test.sh` checks Go formatting and runs the same test suite that CI runs on the supported operating-system matrix for pull requests and manual dispatches:

- `darwin/arm64`
- `linux/arm64`
- `windows/amd64`

## Builds

Build the host binary:

```bash
bin/build.sh
```

Build explicit release archives:

```bash
bin/build.sh --target darwin/arm64
bin/build.sh --target linux/arm64
bin/build.sh --target windows/amd64
bin/build.sh --target all
```

Install locally:

```bash
bin/install.sh
```

Notes:

- `bin/build.sh` runs `bin/test.sh` before building the CLI, unless `--no-test` is passed
- `bin/build.sh` keeps the host binary at `build/refactorlah` and writes release archive directories under `build/dist/refactorlah_<goos>-<goarch>/`
- `bin/install.sh` runs `bin/build.sh --target host`, so it also runs the test suite first
- local install copies the host binary to `refactorlah` in the install directory, so the command does not depend on the repository checkout after install
- built-in PHP/Python support is compiled through cgo; cross-target builds require a C compiler for the requested `GOOS/GOARCH`, or should be run on a matching CI/runner OS

## Releases

Build and publish release archives through GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow can also be started manually from GitHub Actions. Tag pushes publish GitHub releases; manual dispatches build artefacts only.

Supported release targets are:

- `darwin/arm64`
- `linux/arm64`
- `windows/amd64`

Release runs repeat the supported operating-system test matrix before building publishable archives. Release builds are intentionally tag/manual only, so normal pull requests do not spend release-build minutes.

## Adapter Structure

Native adapter source lives under `internal/adapters`:

- `internal/adapters/php`, `python`, and `golang` contain language-specific analysers and rules
- `internal/adapters/staticimports` handles shared deterministic static asset import rewrites until fuller frontend adapters exist
- `internal/adapters/registry` wires built-in adapters into the CLI
- `internal/adapters/contract` contains shared semantic result types
- `internal/parsing/treesitter` contains parser infrastructure, not adapter logic

Adapter guidance lives in [internal/adapters/README.md](internal/adapters/README.md).
