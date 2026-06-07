# General Notes

This page collects behaviour that is important in real projects but too detailed for the README.

## Path Resolution

`refactorlah` keeps all internal paths project-relative, but command-line paths may be written from the directory where you run the command.

Resolution rules for `refactorlah move`:

- Absolute paths are accepted if they stay inside the project root.
- Relative old paths are resolved from the current working directory first when that path exists.
- If the current-directory path does not exist, the old path falls back to the project root.
- The new path uses the same base as the old path.
- If the old path exists in both places, the command fails as ambiguous.

For example, from a monorepo root:

```bash
refactorlah move platform/src/Old.php platform/src/New.php
refactorlah move collector/src/collector/old.py collector/src/collector/new.py
```

From inside `platform/`, the shorter form is also valid:

```bash
refactorlah move src/Old.php src/New.php
```

If both `src/Old.php` and `platform/src/Old.php` exist, `refactorlah` does not guess. Use the explicit project-relative path:

```bash
refactorlah move platform/src/Old.php platform/src/New.php
```

The same base selection applies to wildcard moves.

## Scan Scope

`.refactorlah.json` uses PHPStan-style include/exclude patterns for refactoring scope.

- `exclude` means `refactorlah` should not move, semantically rewrite, or report semantic warnings for matching paths.
- `include` overrides `exclude` for exact exceptions.
- Explicitly requested moves inside excluded paths fail before writing.

This is intentionally strict. Fixture and stub directories often contain namespaces or package names that are only valid in-place; moving them without semantic rewrites would break them, while rewriting them would defeat the purpose of excluding them.
