# Technical Backlog

These are known follow-ups that should not be lost while the native adapters mature.

## Adapter Scanning

- Keep broad project file discovery behind `internal/adapters/scan.Index`.
- Strengthen impact planning so adapters declare candidate extensions and needles before reading files.
- Add more performance-shape tests around large irrelevant corpora and mixed-adapter moves.

## PHP Adapter

- Split `internal/adapters/php/analyzer.go` into smaller collectors for PHP symbols, PHP references, Symfony/Twig paths, Symfony config, static imports, and semantic warnings.
- Replace remaining repeated map/slice plumbing with explicit value objects and collections where it improves refactorability.
- Revisit Composer root handling for monorepos with multiple Composer projects and mixed-language moves.

## Static Imports

- Decide whether `staticimports` should stay shared or become a first-class asset adapter.
- Expand static import coverage for JavaScript, TypeScript, CSS, and framework entrypoint patterns before adding more rewrite categories.

## Fixtures

- Grow fixture coverage towards realistic refactor sessions rather than only small isolated snippets.
- Keep mixed PHP, Python, and Go scenarios in the CLI tests so adapter composition stays covered.
