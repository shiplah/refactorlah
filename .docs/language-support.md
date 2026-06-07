# Language Support

This page tracks implemented, planned, intentionally skipped, and temporarily missing language support.

Status labels:

- Supported: implemented and covered by tests
- Report-only: detected as a warning, but not rewritten
- Planned: intended, but not implemented yet
- Temporary gap: known missing behaviour that should be closed
- Intentionally ignored: not planned because it would require guessing
- N/A: the concept does not apply to that language

## Shared Capabilities

| Capability | PHP support | Python support | Go support |
| --- | --- | --- | --- |
| Batch and wildcard-expanded moves | Supported through the core move plan | Supported through the core move plan | Supported through the core move plan |
| Scan include/exclude config | Supported through `.refactorlah.json` | Supported through `.refactorlah.json` | Temporary gap |
| File symbol/module mapping | Supported for Composer PSR-4 classes, interfaces, traits, and enums | Supported for importable `.py` modules in detected source roots | Supported as package import-path mappings from `go.mod` |
| Directory moves | Supported for deterministic PHP files | Supported for modules under detected source roots | Supported for package-directory import paths |
| Declaration updates | Supported for PHP namespaces and primary class/interface/trait/enum basename changes | N/A; module names are path-derived | Supported for deterministic package basename and top-level symbol basename changes |
| Import rewrites | Supported for simple `use` imports, short references, and same-namespace clean-up | Supported for absolute imports, `from` imports, safe relative imports, and visible module references | Supported for exact Go import paths |
| Type and code references | Supported for FQCNs, class constants, attributes, property types, parameter types, return types, class-like references, and PHPDoc tags | Supported for qualified module references and exact string annotations | Supported for unaliased package qualifiers, imported symbol selectors, and same-package symbol references when deterministic |
| Config/path references | Supported for selected framework config, asset paths, and static imports | Supported for exact dotted module references in TOML, INI, CFG, YAML, and YML | Temporary gap |
| Dynamic references | Report-only where recognised, otherwise intentionally ignored | Report-only for dynamic imports where recognised | N/A |
| Arbitrary strings and semantic names | Report-only for likely renamed semantic names | Intentionally ignored unless they are exact supported config or annotation references | Intentionally ignored |
| Unusual project layouts | Temporary gap where Composer/Twig config cannot prove mappings | Temporary gap for unusual source-root/package layouts | Temporary gap outside ordinary `go.mod` modules |
| Validation | Supported through Composer, PHPStan, Psalm, and optional Composer tests | Supported through configured Ruff, MyPy, and optional Pytest | Temporary gap |

## PHP Ecosystem Coverage

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

## Python Ecosystem Coverage

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

## Go Ecosystem Coverage

| Area | Status | Notes |
| --- | --- | --- |
| Go modules | Supported | Reads `go.mod` to derive moved package import paths. |
| Full package directory moves | Supported | Semantic Go rewrites run only when all `.go` files in the source package directory move to one target package directory. Partial package moves are warned and skipped semantically. |
| Import paths | Supported | Rewrites exact string import specs in Go files when a moved package directory changes import path. |
| Package declarations | Supported | Rewrites `package oldpkg` to `package newpkg` only when the old package name matches the old directory basename, including `_test` packages. Custom package names are preserved. |
| Package qualifiers | Supported | Rewrites unaliased references such as `oldpkg.Build()` to `newpkg.Build()` when the import path and package rename are deterministic. Explicit import aliases and locally shadowed names are preserved. |
| Go symbols and references | Supported | Rewrites top-level type, function, const, and var declarations when the old/new file basenames map deterministically to the old/new symbol names. Same-package references are resolved with Go type information; imported selectors such as `models.OldThing` preserve aliases and become `models.NewThing`. |
| Arbitrary Go symbol renames | Planned | Renaming symbols that do not match the moved file basename is not implemented yet. Constructors, methods, test-name conventions, and semantic name families are intentionally not guessed. |
| Go config and generated files | Temporary gap | Go-specific config rewrites are not implemented yet. |
