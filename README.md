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
- builds a source-checkout-independent `build/` bundle
- copies the bundle into the install directory
- installs `refactorlah` via symlink in `~/.local/bin` by default

To use a different install directory:

```bash
bin/install.sh ~/bin
```

Source installs require Go with cgo support so the native language parsers are compiled into the CLI. The installed bundle does not depend on the source checkout at runtime, and PHP/Python refactors do not require PHP, Composer, or Python on the target machine.

The full development test suite still exercises the legacy PHP and Python adapter packages, so `bin/test.sh` requires PHP 8.2+, Composer, and Python 3.11+.

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
- Go package moves with deterministic import-path rewrites
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

## Language Support Status

Status labels:

- Supported: implemented and covered by tests
- Report-only: detected as a warning, but not rewritten
- Planned: intended, but not implemented yet
- Temporary gap: known missing behaviour that should be closed
- Intentionally ignored: not planned because it would require guessing
- N/A: the concept does not apply to that language

### Shared Capabilities

| Capability | PHP support | Python support | Go support |
| --- | --- | --- | --- |
| Batch and wildcard-expanded moves | Supported through the core move plan | Supported through the core move plan | Supported through the core move plan |
| Scan include/exclude config | Supported through `.refactorlah.json` | Supported through `.refactorlah.json` | N/A for current import-only scans |
| File symbol/module mapping | Supported for Composer PSR-4 classes, interfaces, traits, and enums | Supported for importable `.py` modules in detected source roots | Supported as package import-path mappings from `go.mod` |
| Directory moves | Supported for deterministic PHP files | Supported for modules under detected source roots | Supported for package-directory import paths |
| Declaration updates | Supported for PHP namespaces and primary class/interface/trait/enum basename changes | N/A; module names are path-derived | N/A; package declarations are not renamed |
| Import rewrites | Supported for simple `use` imports, short references, and same-namespace clean-up | Supported for absolute imports, `from` imports, safe relative imports, and visible module references | Supported for exact Go import paths |
| Type and code references | Supported for FQCNs, class constants, attributes, property types, parameter types, return types, class-like references, and PHPDoc tags | Supported for qualified module references and exact string annotations | N/A for current import-only support |
| Config/path references | Supported for selected framework config, asset paths, and static imports | Supported for exact dotted module references in TOML, INI, CFG, YAML, and YML | Temporary gap |
| Dynamic references | Report-only where recognised, otherwise intentionally ignored | Report-only for dynamic imports where recognised | N/A |
| Arbitrary strings and semantic names | Report-only for likely renamed semantic names | Intentionally ignored unless they are exact supported config or annotation references | Intentionally ignored |
| Unusual project layouts | Temporary gap where Composer/Twig config cannot prove mappings | Temporary gap for unusual source-root/package layouts | Temporary gap outside ordinary `go.mod` modules |
| Validation | Supported through Composer, PHPStan, Psalm, and optional Composer tests | Supported through configured Ruff, MyPy, and optional Pytest | Temporary gap |

### PHP Ecosystem Coverage

| Area | Status | Notes |
| --- | --- | --- |
| Composer/PSR-4 | Supported | Required for deterministic PHP symbol mappings. Unknown PHP files are warned and skipped semantically. |
| PHP namespaces and imports | Supported | Includes namespace updates, primary symbol basename changes, simple `use` imports, short references, and same-namespace import clean-up. |
| PHP type and doc references | Supported | Covers FQCNs, class constants, attributes, property types, parameter types, return types, class-like references, and PHPDoc tags. |
| Symfony/Twig | Supported | Covers static Twig references, `render()`, `renderView()`, `#[Template]`, Twig component attributes, and selected YAML/PHP config. |
| Static frontend imports | Supported | Exact JS/TS/CSS-style import specifiers for moved assets are rewritten when the new relative path is deterministic. |
| Group `use` statements | Report-only | Group imports that reference moved symbols are detected and warned, but still not rewritten. |
| Laravel/Blade | Planned | Not implemented yet. It should live as Laravel-specific coverage, not as generic PHP behaviour. |
| Dynamic PHP/Twig references | Intentionally ignored | Concatenated class names, variable template names, mixed Twig fallback arrays, and runtime-dependent references are not rewritten. |

### Python Ecosystem Coverage

| Area | Status | Notes |
| --- | --- | --- |
| Python modules | Supported | Moves are mapped to module names from detected source roots such as `src/` layouts and project packages. |
| Absolute and relative imports | Supported | Covers `import old.module`, `from old.module import Name`, safe relative module imports, and visible short module references. |
| String annotations | Supported | Exact module references inside string annotations are rewritten. |
| Config dotted paths | Supported | Exact dotted module references in TOML, INI, CFG, YAML, and YML are rewritten. |
| Comments, docstrings, and arbitrary strings | Intentionally ignored | These are usually prose, examples, or fixtures rather than executable references. |
| Dynamic imports | Report-only | Runtime imports such as `importlib.import_module(...)` are warned where recognised, not rewritten. |
| Jinja templates | Planned | Not implemented yet. It should be framework/template coverage, not generic Python import behaviour. |
| Django templates | Planned | Not implemented yet. Django-specific conventions should be modelled separately from generic Python module moves. |

### Go Ecosystem Coverage

| Area | Status | Notes |
| --- | --- | --- |
| Go modules | Supported | Reads `go.mod` to derive moved package import paths. |
| Import paths | Supported | Rewrites exact string import specs in Go files when a moved `.go` file changes package directory. |
| Package declarations | Intentionally ignored | Moving a package directory does not require changing `package` names, and guessing package renames would be unsafe. |
| Go symbols and references | Planned | Type/function/identifier moves are not implemented yet. |
| Go config and generated files | Temporary gap | Current Go support is intentionally narrow and import-path focused. |

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

Run the full test suite:

```bash
bin/test.sh
```

Run only the native Go test suite:

```bash
bin/test.sh --go-only
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

- `bin/build.sh` runs `bin/test.sh --go-only` before building the native CLI
- `bin/install.sh` runs `bin/build.sh`, so it also runs the native Go test suite first
- the build output is a source-checkout-independent bundle under `build/`
- local install copies that bundle into the install directory, so the command does not depend on the repository checkout after install

## Status

This is a safe working foundation, not a complete universal refactoring engine. It is designed to reduce fragile manual refactor workflows, especially for agents, while staying conservative about what it rewrites.
