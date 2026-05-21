# RefactorLah -- Refactor tooling for AI Agents

A conservative refactoring CLI for humans and AI agents. It is built for the common case where renaming or moving code is more than a filesystem operation, but still should not require a chain of separate tool calls. Instead of manually combining `git mv`, namespace edits, import clean-up, reference updates, and follow-up validation, you run one command and get a deterministic result.

`refactorlah` does **not** try to be a universal refactor engine. It rewrites only references it can prove from project configuration, and warns on anything uncertain.

## Install

Until release archives are published, install from the repository:

```bash
git clone git@github.com:NickSdot/refactorlah.git
cd refactorlah
bin/install.sh
```

What `bin/install.sh` does:

- runs `bin/test.sh`
- builds a self-contained `build/` bundle
- installs `refactorlah` via symlink in `~/.local/bin` by default

To use a different install directory:

```bash
bin/install.sh ~/bin
```

## Usage

Apply a move:

```bash
refactorlah move app/Services/Billing app/Domain/Billing
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
- deterministic template-path rewrites where project configuration makes them provable
- text and JSON reporting
- optional post-apply validation

Conservative skips in v1:

- dynamic references
- non-deterministic string rewrites
- group `use` rewrites
- unsupported language- or framework-specific cases

## Commands

### `move`

Move files or directories and rewrite deterministic references.

Examples:

```bash
refactorlah move app/Services/Billing app/Domain/Billing
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
- `--no-adapters`
- `--format=text`
- `--format=json`
- `--no-validation`
- `--run-tests`

Validation:

- `composer dump-autoload`
- `vendor/bin/phpstan`
- `vendor/bin/psalm`
- `composer test` when `--run-tests` is passed and the target project defines it

Use `--no-validation` to skip validation.

Wildcard support:

- `*` is supported in both old and new paths
- old and new must contain the same number of `*` placeholders
- each `*` matches within a single path segment
- `**` is not supported

## Contributing

Run the full test suite:

```bash
bin/test.sh
```

Build the release bundle:

```bash
bin/build.sh
```

Install locally:

```bash
bin/install.sh
```

Notes:

- `bin/build.sh` runs `bin/test.sh` before building
- `bin/install.sh` runs `bin/build.sh`, so it also runs the test suite first
- the build output is a self-contained bundle under `build/`

## Status

This is a safe working foundation, not a complete universal refactoring engine. It is designed to reduce fragile manual refactor workflows, especially for agents, while staying conservative about what it rewrites.
