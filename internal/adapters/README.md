# Native Adapters

`internal/adapters` contains the semantic analysers that are compiled into the `refactorlah` binary.

Adapters inspect a planned move and return deterministic semantic proposals through `internal/adapters/contract`. They must not move files, write files, run validation, inspect CLI flags, or print output.

## Layout

- `contract`: shared result types returned by adapters.
- `registry`: built-in adapter registration.
- `php`: PHP, Composer/PSR-4, Symfony, and Twig support.
- `python`: Python module and import support.
- `golang`: Go package and deterministic top-level symbol support.
- `staticimports`: shared exact static asset import rewrites.
- `shared`: small adapter helpers used by more than one adapter.

Parser infrastructure belongs in `internal/parsing`, not in an adapter package.

## Rules

Add new rewrite behaviour as a focused rule with focused tests. A rule should accept explicit input, return replacement proposals or warnings, and avoid hidden global state.

Keep feature parity in mind: when two languages share a concept, such as import rewrites or config include/exclude handling, the behaviour should be aligned unless the language semantics make that unsafe.

Do not reintroduce external runtime adapter packages. The shipped product direction is one native Go CLI with built-in semantic adapters.
