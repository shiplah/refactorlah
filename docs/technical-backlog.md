# Technical Backlog

These are known follow-ups that should not be lost while the native adapters mature.

## Adapter Scanning

- Measure Go package/symbol move scans before narrowing them further. Go currently parses `.go` candidates broadly inside the module so package declarations, imports, local symbols, and external qualifiers stay consistent.
- Add larger fixture-backed performance tests if real projects still show slow mixed-adapter moves after candidate query filtering.

## PHP Adapter

- Continue replacing lower-level Symfony/Twig map and slice plumbing with explicit value objects when it removes real duplication or clarifies ownership.
- Keep `internal/adapters/php/analyzer.go` orchestration-only as new PHP behaviour is added.

## Static Imports

- Decide whether `staticimports` should stay shared or become a first-class asset adapter.
- Expand static import coverage for JavaScript, TypeScript, CSS, and framework entrypoint patterns before adding more rewrite categories.

## Fixtures

- Grow fixture coverage towards realistic refactor sessions rather than only small isolated snippets.
- Keep mixed PHP, Python, and Go scenarios in the CLI tests so adapter composition stays covered.
