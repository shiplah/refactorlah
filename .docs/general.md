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

`.refactorlah.json` uses PHPStan-style include/exclude patterns for semantic scanning.

- `exclude` means `refactorlah` should not semantically touch matching paths.
- `include` overrides `exclude` for exact exceptions.
- The core still validates requested move paths separately from semantic scanning.

The exact policy for explicit moves inside excluded paths is intentionally strict and should avoid breaking fixture or stub directories.
