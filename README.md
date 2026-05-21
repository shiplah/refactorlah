# refactorlah

`refactorlah` is a conservative refactoring CLI for moving files and directories, then updating only deterministic references that can be proven from project configuration.

It is designed to help humans and coding agents avoid brittle workflows like:

- moving files by hand
- fixing PHP namespaces manually
- updating `use` statements one by one
- chasing stale class references
- fixing static Twig template paths manually

It does **not** try to be a magical universal refactorer.

## Principles

- Default to action.
- Never guess.
- Rewrite only deterministic references.
- Warn on dynamic or ambiguous cases.
- Keep language-specific logic in adapters.
- Let the core CLI own planning, moving, validation, and reporting.

## Current scope

Today the repo contains:

- a Go core CLI: `refactorlah`
- a PHP adapter: `refactorlah-php`

The first implemented target is Composer/PSR-4 PHP projects with conservative Twig support.

Implemented PHP rewrite categories:

- namespace declarations in moved PHP files
- simple `use` statements
- exact fully-qualified class references
- `::class` references
- `@var`, `@param`, `@return`, `@throws` docblocks
- typed properties
- parameter types
- return types
- attribute class references where static and exact

Implemented Twig/template rewrite categories:

- `{% include %}`
- `{{ include(...) }}`
- `{% extends %}`
- `{% embed %}`
- `{% use %}`
- `{% import %}`
- `{% from %}`
- Symfony `render()` / `renderView()`
- `#[Template(...)]`
- exact YAML `template: '...'` values

Skipped conservatively in v1:

- dynamic class strings
- dynamic Twig paths
- group `use` rewrites
- Blade rewrites
- non-deterministic string replacements

## Build

If you just want a runnable artifact, use the build script:

```bash
bin/build.sh
```

That creates a self-contained `build/` directory containing:

- `build/refactorlah`
- `build/libexec/refactorlah-php/...`
- `build/README.txt`

The built CLI automatically discovers the bundled PHP adapter next to itself, so you do not need to set `REFACTORLAH_PHP_ADAPTER` for normal use of the build output.

Important:

- the built artifact does not depend on this source repository at runtime
- it still relies on PHP being available on the machine when a PHP refactor is executed
- the current architecture is a bundled CLI plus bundled adapter, not a single static binary containing a PHP runtime

## Install locally

If you want a normal shell command after building:

```bash
bin/install.sh
```

`bin/install.sh` runs `bin/build.sh` automatically before installing the symlink.

By default this installs a symlink at:

```bash
~/.local/bin/refactorlah
```

You can also choose a different install directory:

```bash
bin/install.sh ~/bin
```

After that, normal usage becomes:

```bash
cd ~/Code/example/project
refactorlah old/path new/path
```

## Development setup

If you are working on the adapter itself, the manual setup is still useful:

```bash
go build -o refactorlah ./cmd/refactorlah
cd adapters/php
composer install
```

During development from the source tree, the easiest adapter override is:

```bash
export REFACTORLAH_PHP_ADAPTER="$PWD/adapters/php/bin/refactorlah-php"
```

## Basic usage

Default mode applies changes:

```bash
./build/refactorlah app/Services/Billing app/Domain/Billing
```

or, if you ran `bin/install.sh`:

```bash
refactorlah app/Services/Billing app/Domain/Billing
```

Preview only:

```bash
./build/refactorlah app/Services/Billing app/Domain/Billing --dry-run
```

Move a single PHP file:

```bash
./build/refactorlah app/Services/Billing/InvoiceService.php app/Domain/Billing/InvoiceService.php --dry-run
```

Move a Twig directory:

```bash
./build/refactorlah templates/admin templates/backoffice --dry-run
```

Machine-readable output:

```bash
./build/refactorlah app/Services/Billing app/Domain/Billing --format=json
```

Disable adapters and perform filesystem/git moves only:

```bash
./build/refactorlah app/Services/Billing app/Domain/Billing --dry-run --no-adapters
```

## Options

- `--dry-run`
- `--require-clean`
- `--require-git`
- `--no-adapters`
- `--format=text`
- `--format=json`
- `--no-validation`
- `--run-tests`

Notes:

- If `--dry-run` is not passed, changes are applied.
- `--require-clean` restores the old “clean working tree only” safety check.
- `--require-git` restores the old “git repository required” safety check.

## Adapter behavior

The core auto-detects adapters.

For the PHP adapter, detection is based on signals like:

- `composer.json` exists
- moved paths include `.php`
- moved paths include `.twig`
- Composer has PSR-4 mappings
- a `templates/` directory exists

If PHP/Twig analysis is relevant but the PHP adapter is unavailable:

- dry-run warns and skips semantic rewrites
- apply mode fails unless `--no-adapters` is passed

## Git behavior

Inside a git repository:

- tracked files move via `git mv`
- untracked files fall back to native filesystem rename
- directory moves expand into per-file moves

Outside git:

- native filesystem moves are used
- `--require-git` turns this back into an error

## Validation

After apply, the core can run:

- `composer dump-autoload`
- `vendor/bin/phpstan`
- `vendor/bin/psalm`
- `composer test` when `--run-tests` is passed and the script exists

Validation can be disabled with:

```bash
--no-validation
```

## Example output

Text mode reports:

- files to move
- whether each file is tracked
- which adapter was auto-detected
- derived PHP symbol mappings
- derived Twig path mappings
- files to edit
- replacement worker counts
- warnings
- validation results

## Tests

Go tests:

```bash
GOCACHE=/private/tmp/refactorlah-gocache go test ./...
```

PHP adapter tests:

```bash
cd adapters/php
composer test
```

## Status

This is a safe working foundation, not a complete universal refactoring engine.

The tool intentionally prefers:

- skipping risky rewrites
- emitting warnings
- failing before unsafe writes

over making aggressive guesses.
