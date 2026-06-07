# Usage

## Commands

### Move Command

Move one file or directory:

```bash
refactorlah move app/Services/Billing app/Domain/Billing
refactorlah move app/Services/Billing/InvoiceService.php app/Domain/Billing/InvoiceService.php
```

Preview without writing:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --dry
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

Write JSON output:

```bash
refactorlah move app/Services/Billing app/Domain/Billing --format=json
```

#### Wildcards

- `*` is supported in both old and new paths.
- Old and new paths must contain the same number of `*` placeholders.
- Each `*` matches within a single path segment.
- `**` is not supported.

### Options

- `--dry`
- `--require-clean-worktree`
- `--use-list`
- `--use-file`
- `--format=text`
- `--format=json`
- `--no-validation`
- `--run-tests`

## Path Resolution

`refactorlah` stores paths internally as project-relative paths, but command-line paths may be written from the directory where you run the command.

Resolution rules:

- Absolute paths are accepted when they stay inside the project root.
- Relative old paths are resolved from the current working directory first when that path exists.
- If the current-directory path does not exist, the old path falls back to the project root.
- The new path uses the same base as the old path.
- If the old path exists in both places, the command fails as ambiguous.

From a repository root:

```bash
refactorlah move packages/billing/src/Old.php packages/billing/src/New.php
refactorlah move apps/worker/src/app/old.py apps/worker/src/app/new.py
```

From inside `packages/billing/`, the shorter form is also valid:

```bash
refactorlah move src/Old.php src/New.php
```

If both `src/Old.php` and `packages/billing/src/Old.php` exist, `refactorlah` does not guess. Use the explicit project-relative path:

```bash
refactorlah move packages/billing/src/Old.php packages/billing/src/New.php
```

The same base selection applies to wildcard moves.

## Configuration

### Scan Scope

`.refactorlah.json` uses PHPStan-style include/exclude patterns for refactoring scope:

```json
{
  "exclude": [
    "local/phpstan/tests/fixtures/**"
  ],
  "include": []
}
```

- `exclude` means `refactorlah` should not move, semantically rewrite, or report semantic warnings for matching paths.
- `include` overrides `exclude` for explicit exceptions.
- Explicitly requested moves inside excluded paths fail before writing.

This is intentionally strict. Fixture and stub directories often contain namespaces or package names that are only valid in-place; moving them without semantic rewrites would break them, while rewriting them would defeat the purpose of excluding them.

### Validation

After applying a move, `refactorlah` runs internal replacement safety checks and any available cheap language sanity checks proposed by the relevant adapters, such as PHP linting, Python byte-compilation, Composer autoload generation, or Go builds.

Project-specific quality gates are configured explicitly in `.refactorlah.json`:

```json
{
  "checks": [
    ["composer", "stan"],
    ["ruff", "check", "."]
  ],
  "tests": [
    ["composer", "test"],
    ["go", "test", "./..."]
  ]
}
```

Configured `checks` run after apply. Configured `tests` run only with `--run-tests`.

Use `--no-validation` to skip external sanity checks and configured commands. Replacement conflict validation still always runs before writing.

Language-specific sanity checks are tracked in [features.md](features.md).
