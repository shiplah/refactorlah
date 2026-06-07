# RefactorLah -- deterministic refactoring for agents

A conservative refactoring CLI for humans and AI agents. It handles the common case where moving code is more than a filesystem operation, but still should not require a chain of separate tool calls.

Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, run one command and get a deterministic result.

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

The installed command is source-checkout-independent. PHP and Python refactors do not require PHP, Composer, or Python on the target machine.

To use a different install directory:

```bash
bin/install.sh ~/bin
```

Source installs require Go with cgo support so the native language parsers can be compiled into the CLI.

## Command

### `move`

Move files or directories and rewrite deterministic references.

```bash
refactorlah move app/Services/Billing app/Domain/Billing
```

Preview without writing:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --dry
```

Move one file:

```bash
refactorlah move app/Services/Billing/InvoiceService.php app/Domain/Billing/InvoiceService.php
```

Move a Python module:

```bash
refactorlah move src/app/services/billing.py src/app/domain/billing.py
```

Move multiple pairs inline:

```bash
refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php
```

Move multiple pairs from a file:

```bash
refactorlah move --use-file moves.txt
```

Example `moves.txt`:

```text
app/Foo.php,app/Bar.php
tests/A.php,tests/B.php
```

Move matching files with wildcards:

```bash
refactorlah move 'src/Old/*Worker.php' 'src/New/*Rule.php'
```

JSON output:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --format=json
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

Wildcard rules:

- `*` is supported in both old and new paths
- old and new must contain the same number of `*` placeholders
- each `*` matches within a single path segment
- `**` is not supported

## Scope

Core behaviour:

- moves files or directories
- keeps tracked moves in Git when possible
- falls back cleanly for untracked files
- validates replacements before writing
- warns on dynamic or uncertain cases instead of guessing

Implemented support:

- PHP projects with Composer PSR-4 namespace and class reference rewrites
- Python module moves with deterministic import, string annotation, config dotted-path, and qualified module reference rewrites
- Go package moves and deterministic top-level Go symbol renames
- Symfony/Twig template-path rewrites where project configuration makes them provable
- static asset import rewrites where the new relative path is deterministic
- text and JSON reporting
- optional post-apply validation

Conservative skips:

- dynamic references
- non-deterministic string rewrites
- Python docstrings and arbitrary strings
- Python dynamic imports, such as `importlib.import_module(...)`
- group `use` rewrites
- unsupported language- or framework-specific cases

See [.docs/language-support.md](.docs/language-support.md) for the detailed support matrix and known gaps.

## Configuration

Projects may add `.refactorlah.json` at the command's working directory or up to three directory levels below it to exclude semantic scans for generated, fixture, or stub files:

```json
{
  "exclude": [
    "local/phpstan/tests/fixtures/**"
  ],
  "include": []
}
```

`include` entries override `exclude` entries. The core still plans requested moves; this config only limits semantic rewrites and warnings.

## Validation

After applying a move, `refactorlah` may run available project validation:

- `composer dump-autoload`
- `vendor/bin/phpstan`
- `vendor/bin/psalm`
- `composer test` when `--run-tests` is passed and the target project defines it
- configured Python `ruff` and `mypy`
- configured Python `pytest` when `--run-tests` is passed and available

Use `--no-validation` to skip validation.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for tests, builds, release workflow, and adapter structure.

## Status

This is a safe working foundation, not a complete universal refactoring engine. It is designed to reduce fragile manual refactor workflows, especially for agents, while staying conservative about what it rewrites.
