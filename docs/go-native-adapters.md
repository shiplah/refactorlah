# Go-Native Adapter Concept

`refactorlah` should move towards first-party language adapters implemented in Go and shipped with the main binary. The existing PHP and Python adapters are useful prototypes, but they leak runtime requirements into normal usage. A release user should not need PHP, Python, Composer, Node, or similar tools unless they explicitly run optional validation for that ecosystem.

## Goal

- Ship one normal `refactorlah` binary for each supported platform.
- Keep adapter behaviour deterministic and conservative.
- Share common move, path, configuration, discovery, reporting, and edit-safety logic across languages.
- Keep language-specific logic focused on parsing, symbol discovery, import/reference resolution, and framework-specific path rules.
- Keep external adapters possible later, but do not make them the default release path.

## Tree-sitter Assessment

Tree-sitter is a good default candidate for parsing many source languages because it can be embedded, supports many grammars, and gives us byte-ranged syntax nodes that fit our replacement protocol well.

It is not a complete refactoring engine:

- it knows syntax, not project semantics
- it does not know Composer PSR-4 roots, Python source roots, package manifests, or framework aliases
- it does not resolve imports by itself
- it does not decide whether a rewrite is safe
- it does not replace our move planner, config index, exclusion handling, reporting, or edit validator

So the right model is not "tree-sitter replaces adapters". The model is "Go-native adapters use a parser backend, with tree-sitter as the likely default parser backend where appropriate".

## Spike Findings

The first Go spike uses the official Tree-sitter Go binding with the official PHP and Python grammar modules. It confirms that:

- PHP parsing exposes useful refactor nodes such as `namespace_definition`, `namespace_use_declaration`, `class_declaration`, `namespace_name`, and `qualified_name`
- Python parsing exposes useful refactor nodes such as `import_statement`, `import_from_statement`, `dotted_name`, and `relative_import`
- both grammars provide byte ranges that map directly back to the original source text
- the official PHP and Python Go grammar bindings are cgo-based
- `CGO_ENABLED=0` cannot build those grammar packages

That means Tree-sitter remains a good parser candidate, but adopting it for first-party release builds also means owning cgo-capable builds for every supported platform unless we choose a different parser backend.

## Parser Options

| Option | Strengths | Weaknesses | Recommendation |
| --- | --- | --- | --- |
| Tree-sitter grammars | Broad language coverage, byte offsets, concrete syntax trees, embeddable parser model | Adds parser grammar management; syntax-only; may need cgo or generated grammar packaging decisions | Default candidate for PHP, Python, JavaScript, TypeScript, CSS, HTML, Twig-like syntactic scanning |
| Go standard library parsers | Excellent for Go, no extra runtime, typed AST | Only helps Go and a few data formats | Use for a future Go adapter instead of tree-sitter where it gives better semantics |
| Dedicated Go parser libraries | Can expose richer language-specific ASTs | Quality and maintenance vary per language; may fragment adapter architecture | Consider per language only after a spike proves it is clearly better |
| ANTLR or generated parsers | Mature parser generator ecosystem | Grammar/runtime management, less common in Go refactoring tooling, still syntax-only | Not the first choice |
| Regex/token scanners | Simple and fast for narrow config/path cases | Easy to corrupt overlapping names and miss syntax boundaries | Keep only for tightly scoped non-code strings and config formats |
| Language-native adapters | Best ecosystem fidelity in prototypes | Requires external runtimes and package managers on user machines | Keep as reference/prototype path, not release default |
| LSP/indexer integration | Potentially rich semantic data | External tools, project-specific setup, slower and less deterministic | Optional future validation/reporting, not core rewriting |

## Recommended Architecture

Create a Go-native adapter layer behind the existing adapter protocol shape, but without process execution for first-party adapters.

Suggested packages:

```text
internal/languages/
  adapter.go
  registry.go
  parser/
    parser.go
    treesitter/
  symbols/
    mapping.go
    occurrence.go
    import_model.go
  rules/
    rule.go
    registry.go

internal/languages/php/
  adapter.go
  composer.go
  psr4.go
  parser.go
  references.go
  rules/
  symfony/
  laravel/

internal/languages/python/
  adapter.go
  roots.go
  parser.go
  imports.go
  references.go
  rules/
  django/
  jinja/
```

The existing process-based adapter protocol can remain for optional third-party adapters, but first-party PHP/Python should become in-process Go adapters.

## Shared Infrastructure

These concerns should be shared across language adapters:

- config discovery and merged include/exclude rules
- project-relative path normalisation
- move mapping and wildcard expansion
- candidate file collection
- ignored/generated directory pruning
- binary/text detection
- path-reference mapping for static assets
- replacement validation and overlap detection
- token-boundary helpers
- import insertion block helpers where the concept applies
- report aggregation
- validation orchestration
- adapter capability detection

These concerns should stay language-specific:

- source-root and namespace/module derivation
- symbol discovery
- import/export/reference resolution
- class/function/type/member syntax rules
- framework conventions such as Symfony/Twig, Laravel/Blade, Django templates, or Jinja
- language-specific warnings for dynamic references

## Migration Plan

1. Keep current PHP/Python adapters as behaviour references and regression fixtures.
2. Add a small Go-native adapter interface that returns the same mappings, replacements, warnings, and errors the core already understands.
3. Spike tree-sitter parsing on one narrow PHP move:
   - derive a PSR-4 class mapping
   - update namespace declaration
   - update one import
   - update one imported short reference
4. Spike tree-sitter parsing on one narrow Python move:
   - derive module mapping
   - update absolute import
   - update relative import
   - update imported module reference
5. Compare tree-sitter against any language-specific Go parser before committing fully.
6. Move shared concepts out of the current PHP/Python adapter implementations into Go value objects and rule contracts.
7. Port rules incrementally, keeping one regression test per existing PHP/Python behaviour.
8. Remove runtime adapter requirements from normal release builds once first-party coverage is good enough.

## Open Questions

- Do we want cgo-based tree-sitter bindings, or generated/pure packaging where possible?
- How do we want to version and audit grammar dependencies?
- Should first-party adapters be enabled by default while process adapters remain hidden/internal?
- Which current PHP/Python behaviours are required before the native adapters can replace the prototypes?
- Do we support template engines through language adapters, framework sub-adapters, or a shared static-template path adapter with framework-specific rules?

## Working Recommendation

Use Go-native adapters as the target architecture. Use tree-sitter as the leading parser candidate for PHP, Python, JavaScript, TypeScript, CSS, and HTML-like languages, but treat it as a parser backend rather than the refactoring engine. Prefer stronger native Go parsers when they exist and give better semantics, especially for a future Go adapter.
