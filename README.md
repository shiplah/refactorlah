# refactorlah

`refactorlah` is a conservative refactoring CLI for moving files and directories, then updating only deterministic references that can be proven from project configuration.

It is meant to replace brittle workflows like manual file moves, namespace edits, `use` cleanup, and static Twig path updates. It does **not** try to be a universal refactor engine.

## Install

Until release archives are published, install from the repository:

```bash
git clone git@github.com:NickSdot/refactorlah.git
cd refactorlah
bin/install.sh
```

`bin/install.sh`:

- runs `bin/test.sh`
- builds a self-contained `build/` bundle
- installs `refactorlah` as a symlink in `~/.local/bin` by default

To use a different install directory:

```bash
bin/install.sh ~/bin
```

The bundled CLI auto-discovers its bundled PHP adapter. You do not need to set `REFACTORLAH_PHP_ADAPTER` for normal use.

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

Example `moves.txt`:

```text
app/Foo.php,app/Bar.php
tests/A.php,tests/B.php
```

JSON output:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --format=json
```

## Behaviour

- apply is the default; use `--dry` to preview
- only deterministic references are rewritten
- uncertain or dynamic cases are warned and skipped
- tracked files use `git mv`
- untracked files fall back to filesystem rename
- directory moves expand into per-file moves
- adapters are auto-detected

Current implemented scope:

- Composer PSR-4 PHP namespace and class reference rewrites
- conservative Twig template path rewrites
- JSON and text reporting
- optional post-apply validation

Conservative skips in v1:

- dynamic class strings
- dynamic Twig paths
- group `use` rewrites
- Blade rewrites
- non-deterministic string replacements

## Commands

- `move`: move files or directories and rewrite deterministic references

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

## Validation

After apply, the core can run:

- `composer dump-autoload`
- `vendor/bin/phpstan`
- `vendor/bin/psalm`
- `composer test` when `--run-tests` is passed and the script exists

Use `--no-validation` to skip validation.

## Development

Run the full test suite:

```bash
bin/test.sh
```

Build the bundle:

```bash
bin/build.sh
```

`bin/build.sh` runs `bin/test.sh` first and produces:

- `build/refactorlah`
- `build/libexec/refactorlah-php/...`
- `build/README.txt`

## Status

This is a safe working foundation, not a complete universal refactoring engine. It prefers skipping risky rewrites, emitting warnings, and failing before unsafe writes rather than making aggressive guesses.
