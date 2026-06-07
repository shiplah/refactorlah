# Usage

## Move Command

Move files or directories:

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

## Options

- `--dry`
- `--require-clean-worktree`
- `--use-list`
- `--use-file`
- `--format=text`
- `--format=json`
- `--no-validation`
- `--run-tests`

## Wildcards

- `*` is supported in both old and new paths
- old and new must contain the same number of `*` placeholders
- each `*` matches within a single path segment
- `**` is not supported

## Validation

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

Language-specific sanity checks are tracked in [language-support.md](language-support.md).
