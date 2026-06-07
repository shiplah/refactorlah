# Features

This page describes what `refactorlah` rewrites today, where it only reports risk, and where it deliberately refuses to guess. It includes language-specific features, framework-specific rewrites, static asset imports, and shared CLI behaviour. Future ideas live in [backlog.md](backlog.md).

## At A Glance

| Area | What works today | What is not automatic |
| --- | --- | --- |
| PHP | Composer PSR-4 class moves, namespace updates, imports, PHP type references, PHPDoc references, class constants, selected Symfony/Twig references, selected config/path strings | Dynamic class names, unsafe group imports, non-PSR-4 symbols, arbitrary prose strings |
| Python | Module moves from detected source roots, absolute imports, safe relative imports, visible module references, string annotations, exact dotted config references | Dynamic imports, comments/docstrings, unusual source-root layouts |
| Go | `go.mod` package moves, import paths, package declarations, deterministic top-level symbol basename renames, package qualifiers, same-package references | Arbitrary symbol renames, generated/config rewrites, partial package moves, unusual module layouts |
| Static imports | Exact JS/TS/CSS-style import specifiers for moved assets when the new relative path is deterministic | Broader frontend framework conventions and dynamic import expressions |

## PHP

PHP support is based on Composer PSR-4 configuration. If a PHP file is outside known PSR-4 roots, `refactorlah` warns and skips semantic PHP rewrites for that file.

Supported rewrites:

- Updates namespace declarations for moved PHP files.
- Updates primary class, interface, trait, and enum names when the file basename changes deterministically.
- Updates simple `use` imports, short imported references, fully-qualified references, class constants, attributes, property types, parameter types, return types, class-like references, and PHPDoc tags.
- Removes deterministic same-namespace imports after a move.
- Rewrites selected Symfony/Twig references, including static Twig paths, `render()`, `renderView()`, `#[Template]`, Twig component attributes, and selected YAML/PHP config.
- Rewrites exact static frontend import specifiers for moved assets when the new relative path can be proven.

Reported or skipped:

- Group `use` statements are reported, not rewritten, until they can be handled without risking formatting or alias mistakes.
- Dynamic class names, concatenated template names, Twig fallback arrays, runtime-dependent strings, and unrelated prose strings are skipped.
- Laravel/Blade support is not implemented yet and is tracked in [backlog.md](backlog.md).

## Python

Python support is based on importable modules under detected source roots, such as `src/` layouts or package directories.

Supported rewrites:

- Updates absolute imports such as `import old.module` and `from old.module import Name`.
- Updates safe relative imports when the moved module can be resolved deterministically.
- Updates visible module references and exact module references inside string annotations.
- Updates exact dotted module references in TOML, INI, CFG, YAML, and YML files.

Reported or skipped:

- Dynamic imports such as `importlib.import_module(...)` are reported where recognised, not rewritten.
- Comments, docstrings, arbitrary strings, and unusual source-root layouts are skipped unless they match an exact supported reference shape.
- Jinja and Django template support are not implemented yet and are tracked in [backlog.md](backlog.md).

## Go

Go support is based on ordinary `go.mod` modules. It is intentionally conservative because Go files often contain several declarations and file names are conventions rather than source-of-truth symbols.

Supported rewrites:

- Updates exact import strings when a moved package directory changes import path.
- Updates package declarations when the old package name matches the old directory basename, including `_test` packages.
- Updates unaliased package qualifiers such as `oldpkg.Build()` when the package import and package rename are deterministic.
- Preserves explicit import aliases and locally shadowed names.
- Updates top-level type, function, const, and var declarations when the old and new file basenames map deterministically to the old and new symbol names.
- Updates same-package references with Go type information, and imported selectors such as `models.OldThing` while preserving aliases.

Reported or skipped:

- Partial package moves are warned and skipped semantically; move the whole package directory when you want Go semantic rewrites.
- Arbitrary Go symbol renames are not supported. That means `old_file.go -> new_file.go` can rename `OldFile` to `NewFile`, but it will not infer unrelated names, constructors, methods, test-name conventions, or semantic name families.
- Go config/generated-file rewrites are not implemented. Generated files and tool-specific config should be reviewed manually unless covered by exact static import rewriting.
- Unusual Go layouts outside ordinary `go.mod` modules are skipped because the import path cannot be proven safely.

## Static Imports

Static import support is deliberately narrow and language-neutral for now. It rewrites exact relative import specifiers for moved non-PHP assets when the old path can be resolved from the importing file and the new relative path can be computed without guessing.

Examples include CSS imported from JavaScript or TypeScript entrypoints. Broader JavaScript, TypeScript, CSS, and framework-specific behaviour is tracked in [backlog.md](backlog.md).

## Shared Behaviour

These behaviours are owned by the core CLI and apply across adapters where the language-specific concepts exist:

- File moves, directory moves, batch moves, `--use-list`, `--use-file`, and wildcard-expanded moves.
- Git-aware moving, with tracked files moved through Git and untracked files moved through the filesystem.
- Conservative replacement validation before writing, including invalid byte ranges and overlapping edits.
- Text and JSON reports for moves, edits, warnings, and validation.
- `.refactorlah.json` project configuration for scan include/exclude and configured checks/tests.
- Built-in sanity checks after apply: PHP lint and Composer autoload checks where applicable, Python byte compilation, Go `go build ./...`, plus configured project checks and tests.
