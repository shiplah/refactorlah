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

After applying a move, `refactorlah` may run available project validation.

Use `--no-validation` to skip validation. Use `--run-tests` to run supported project test commands where available.

Language-specific validation support is tracked in [language-support.md](language-support.md).
