# RefactorLah -- Refactor tooling for AI Agents

A conservative refactoring CLI for humans and AI agents. It is built for the common case where renaming or moving code is more than a filesystem operation, but still should not require a chain of separate tool calls. Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, you run one command and get a deterministic result.

`refactorlah` does **not** try to be a universal refactor engine. It rewrites only references it can prove from project configuration, and warns on anything uncertain.

> [!WARNING]
> This is currently a hacky pre-alpha experiment. It is useful for dogfooding and careful trials, but you should review its output and keep your project under version control before relying on it.

## Install

Until release archives are published, install from the repository:

```bash
git clone git@github.com:NickSdot/refactorlah.git
cd refactorlah
bin/install.sh
```

What `bin/install.sh` does:

- runs the native Go test suite
- builds a source-checkout-independent host bundle
- copies the bundle into the install directory
- installs `refactorlah` via symlink in `~/.local/bin` by default

To use a different install directory:

```bash
bin/install.sh ~/bin
```

Source installs require Go with cgo support so the native language parsers are compiled into the CLI. The install/build scripts are POSIX-shell compatible and can be run through `sh`, Bash, or Zsh. The installed bundle does not depend on the source checkout at runtime, and PHP/Python refactors do not require PHP, Composer, or Python on the target machine.

## Usage

Apply a move:

```bash
refactorlah move app/Services/Billing app/Domain/Billing
```

Move a Python module:

```bash
refactorlah move src/app/services/billing.py src/app/domain/billing.py
```

Preview only:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --dry
```

Move one file:

```bash
refactorlah move app/Services/Billing/InvoiceService.php app/Domain/Billing/InvoiceService.php
```

Move multiple pairs inline:

```bash
refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php
```

Move multiple pairs from a file:

```bash
refactorlah move --use-file moves.txt
```

Move matching files with wildcards:

```bash
refactorlah move 'src/Old/*Worker.php' 'src/New/*Rule.php'
```

Example `moves.txt`:

```text
app/Foo.php,app/Bar.php
tests/A.php,tests/B.php
```

JSON output:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --format=json
```

## What it does

- moves files or directories
- updates deterministic references that can be proven safely
- keeps tracked moves in Git when possible
- falls back cleanly for untracked files
- validates replacements before writing anything
- warns on dynamic or uncertain cases instead of guessing

Current implemented scope:

- PHP projects with Composer PSR-4 namespace and class reference rewrites
- Python module moves with deterministic import, string annotation, config dotted-path, and qualified module reference rewrites
- Go package moves and deterministic top-level Go symbol renames
- Symfony/Twig template-path rewrites where project configuration makes them provable
- text and JSON reporting
- optional post-apply validation

Project root detection:

- Git root, when available
- otherwise the current working directory

Conservative skips in v1:

- dynamic references
- non-deterministic string rewrites
- Python docstrings and arbitrary strings
- Python dynamic imports, such as `importlib.import_module(...)`
- group `use` rewrites
- unsupported language- or framework-specific cases

## Language Support

Current native support covers deterministic PHP, Python, Go, Symfony/Twig, and static asset import rewrites. See [.docs/language-support.md](.docs/language-support.md) for the full support matrix and known gaps.

## Commands

### `move`

Move files or directories and rewrite deterministic references.

Examples:

```bash
refactorlah move app/Services/Billing app/Domain/Billing
refactorlah move src/app/services/billing.py src/app/domain/billing.py
refactorlah move app/Services/Billing app/Domain/Billing --dry
refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php
refactorlah move --use-file moves.txt
refactorlah move 'src/Old/*Worker.php' 'src/New/*Rule.php'
```

Options:

- `--dry`
- `--require-clean-worktree`
- `--use-list`
- `--use-file`
- `--format=text`
- `--format=json`
- `--no-validation`
- `--run-tests`

Validation:

- `composer dump-autoload`
- `vendor/bin/phpstan`
- `vendor/bin/psalm`
- `composer test` when `--run-tests` is passed and the target project defines it
- configured Python `ruff` and `mypy`, when available
- configured Python `pytest` when `--run-tests` is passed and available

Use `--no-validation` to skip validation.

Wildcard support:

- `*` is supported in both old and new paths
- old and new must contain the same number of `*` placeholders
- each `*` matches within a single path segment
- `**` is not supported

## Configuration

Projects may add `.refactorlah.json` at the command's working directory or up to three directory levels below it to exclude semantic scans for generated, fixture, or stub files. Patterns are slash-normalised, resolved through absolute paths relative to the config file that declares them, deduplicated, and then shared with analysers as project-relative rules:

```json
{
  "exclude": [
    "local/phpstan/tests/fixtures/**"
  ],
  "include": []
}
```

`include` entries override `exclude` entries. The core still plans requested moves; this config only limits semantic rewrites and warnings. The scan policy is shared with every language analyser, so future JavaScript, TypeScript, or CSS support will receive the same include/exclude rules.

## Contributing

Run the test suite:

```bash
bin/test.sh
```

Build the release bundle:

```bash
bin/build.sh
```

Build explicit native release bundles:

```bash
bin/build.sh --target darwin/arm64
bin/build.sh --target linux/arm64
bin/build.sh --target windows/arm64
bin/build.sh --target all
```

Build and publish release archives through GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow can also be started manually from GitHub Actions. It runs
only for `v*` tags or manual dispatches, builds each native target on a matching
GitHub-hosted runner, uploads the archives as workflow artefacts, and publishes
tagged builds as GitHub releases.

Supported release targets are ARM-only:

- `darwin/arm64`
- `linux/arm64`
- `windows/arm64`

CI runs the test suite on the same supported operating-system matrix for pull
requests and manual dispatches. Release runs repeat that matrix before building
publishable archives.

Install locally:

```bash
bin/install.sh
```

Notes:

- `bin/build.sh` runs `bin/test.sh` before building the native CLI, unless `--no-test` is passed
- `bin/build.sh` keeps the host binary at `build/refactorlah` and writes target bundles under `build/dist/refactorlah_<goos>-<goarch>/`
- `bin/install.sh` runs `bin/build.sh --target host`, so it also runs the test suite first
- local install copies the host bundle into the install directory, so the command does not depend on the repository checkout after install
- native PHP/Python support is compiled through cgo; cross-target builds require a C compiler for the requested `GOOS/GOARCH`, or should be run on a matching CI/runner OS
- release builds are intentionally tag/manual only, so normal pushes and pull requests do not spend release-build minutes

Native adapter source lives under `internal/adapters`:

- `internal/adapters/php`, `python`, and `golang` contain language-specific analysers and rules
- `internal/adapters/staticimports` handles shared deterministic static asset import rewrites until fuller frontend adapters exist
- `internal/adapters/registry` wires built-in adapters into the CLI
- `internal/adapters/contract` contains shared semantic result types
- `internal/parsing/treesitter` contains parser infrastructure, not adapter logic

## Status

This is a safe working foundation, not a complete universal refactoring engine. It is designed to reduce fragile manual refactor workflows, especially for agents, while staying conservative about what it rewrites.
