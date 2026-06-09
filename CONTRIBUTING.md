# Contributing

```bash
git clone git@github.com:shiplah/refactorlah.git
cd refactorlah
bin/test.sh
bin/install.sh ~/your-dev-bin
```

Requirements:

- Go
- a working C toolchain for the current native parser bindings

The C toolchain requirement comes from the current tree-sitter bindings. Removing it would mean changing parser bindings or parser strategy, so for now it is part of local development.

## Local Workflow

```bash
bin/format.sh
bin/test.sh
bin/build.sh
bin/install.sh ~/your-dev-bin
```

`bin/format.sh` applies Go formatting. `bin/test.sh` checks formatting and runs the full test suite, so CI fails if formatting is missing. `bin/build.sh` runs tests before building unless `--no-test` is passed. `bin/install.sh` runs the build first and installs the local binary.

## Pull Requests

Before opening a PR:

- run `bin/format.sh`
- run `bin/test.sh`
- add focused regression tests for bug fixes
- keep changes small and easy to review
- avoid mixing unrelated refactors, docs, fixes, and features
- use conventional commit messages where practical

CI runs `bin/test.sh` on macOS, Linux, and Windows.

## Project Map

Useful docs:

- [README.md](README.md): user-facing overview and install instructions
- [.docs/usage.md](.docs/usage.md): command usage and configuration
- [.docs/features.md](.docs/features.md): supported and unsupported behaviour
- [.docs/backlog.md](.docs/backlog.md): known planned gaps
- [internal/adapters/README.md](internal/adapters/README.md): adapter architecture and expectations
- [AGENTS.md](AGENTS.md): repository rules for coding agents

Useful directories:

- `cmd/refactorlah`: CLI entrypoint
- `internal/cli`: command orchestration
- `internal/adapters`: built-in semantic adapters
- `internal/parsing`: parser infrastructure
- `internal/replacements`: replacement validation and application
- `tests/fixtures`: end-to-end fixture projects

## Adapter Changes

Adapter code proposes replacements only. It must not move files, write files, inspect Git state, run validation, or print output.

For adapter work:

1. Add or update a focused rule.
2. Add rule-level tests.
3. Add adapter-level tests for composition.
4. Add fixture or CLI-level coverage for real workflows.
5. Update [.docs/features.md](.docs/features.md) or [.docs/backlog.md](.docs/backlog.md) when support changes.

See [internal/adapters/README.md](internal/adapters/README.md) for the full adapter contract.
