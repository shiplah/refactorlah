# Native Adapters

`internal/adapters` contains the semantic analysers that are compiled into the `refactorlah` binary.

Adapters inspect a planned move and return deterministic semantic proposals through `internal/adapters/contract`. They must not move files, write files, run validation, inspect CLI flags, or print output.

The goal is feature parity where language concepts overlap. If PHP, Python, and Go all have deterministic import rewrites, each adapter should cover that concept with comparable tests. If a concept does not exist for a language, document it as not applicable rather than silently ignoring it.

## Layout

- `contract`: shared result types returned by adapters.
- `registry`: built-in adapter registration.
- `php`: PHP, Composer/PSR-4, Symfony, and Twig support.
- `python`: Python module and import support.
- `golang`: Go package and deterministic top-level symbol support.
- `staticimports`: shared exact static asset import rewrites.
- `shared`: small adapter helpers used by more than one adapter.

Parser infrastructure belongs in `internal/parsing`, not in an adapter package.

Framework or ecosystem-specific behaviour should live below the language adapter that owns the project semantics. For example, Symfony/Twig belongs under `php/symfony/twig`; future Laravel/Blade, Django/templates, or Jinja support should follow the same pattern instead of becoming generic catch-all rules.

## Adapter Contract

An adapter receives the already-built core move plan and scan config through its Go API. It returns:

- symbol mappings for deterministic symbol/module/package moves
- path mappings for deterministic path-like references
- replacement proposals with byte offsets into original file contents
- warnings for suspicious but uncertain references

Adapters must not:

- apply replacements
- move files
- inspect or mutate Git state
- run validation or tests
- read CLI flags directly
- print to stdout or stderr
- depend on external PHP, Python, Composer, or package-manager runtimes

## Rules

Add new rewrite behaviour as a focused rule with focused tests. A rule should accept explicit input, return replacement proposals or warnings, and avoid hidden global state.

Keep feature parity in mind: when two languages share a concept, such as import rewrites or config include/exclude handling, the behaviour should be aligned unless the language semantics make that unsafe.

Each rule should cover exactly one rewrite category. Prefer names that describe the reference type, such as `UseStatementRule`, `RelativeImportRule`, or `PackageQualifierRule`. Avoid broad scanners that mix symbol derivation, file collection, warnings, replacement creation, and framework semantics in one file.

## Adding Or Changing An Adapter

1. Add or update focused rules in the relevant adapter package.
2. Add rule-level tests beside each rule.
3. Add adapter-level tests that prove the rules compose correctly on realistic move plans.
4. Add or update fixture smoke tests when the behaviour affects end-to-end CLI usage.
5. Register new first-party adapters in `registry`.
6. Update the README support matrix when support, planned gaps, or intentionally ignored cases change.
7. Run `bin/test.sh`.

When a new adapter needs parser support, put parser wrappers in `internal/parsing` and keep language semantics in the adapter package.

## Coverage Expectations

For every deterministic rewrite category, tests should cover:

- happy path replacements
- byte offsets into original content
- directory moves where applicable
- basename/symbol rename moves where applicable
- skipped uncertain references with warnings
- configured include/exclude rules
- dynamic or ambiguous references that must not be rewritten
- preservation of existing reference style per occurrence

Regression tests should be added for every real-world bug class before or alongside the fix. If a bug cannot be reproduced in a small rule-level test, add an adapter-level or fixture-level test instead.

## Shared Behaviour

Prefer shared helpers only for language-neutral mechanics such as converting replacement values into contract results. If a helper starts knowing about imports, namespaces, packages, source roots, or framework conventions, it belongs in an adapter-specific package.

Do not reintroduce external runtime adapter packages. The shipped product direction is one native Go CLI with built-in semantic adapters.
