# AGENTS

This repository is a conservative refactoring tool. Fresh contributors should optimise for safety and determinism over breadth.

## Product rules

- `refactorlah` must rewrite only references it can prove from project configuration.
- If a rewrite is uncertain, skip it and emit a warning instead.
- Do not add “helpful” inferred moves for related files the user did not ask to move.
- Preserve the codebase's existing reference style per occurrence:
  - imported short name stays short
  - explicit fully-qualified name stays fully-qualified
- Namespace and import clean-up is allowed only when it is deterministic.
- The current shipped native language support is PHP, Python, deterministic Go package-move rewriting, and deterministic Go top-level symbol renames. Keep shared behaviour aligned where the language concepts match.

## Architecture

- Keep the Go core responsible for planning, moving, validation, reporting, and applying edits.
- Keep adapters responsible for analysis and replacement proposals only.
- Adapters must not write files.
- Native adapter code lives under `internal/adapters/<adapter>`.
- Shared semantic result types live under `internal/adapters/contract`.
- Built-in adapter registration lives under `internal/adapters/registry`.
- Parser infrastructure lives under `internal/parsing`; do not put adapter rules there.
- Do not reintroduce top-level runtime adapter packages or external adapter process invocation.
- New rewrite behaviour should normally be added as a dedicated rule, not folded into a catch-all scanner.
- Prefer explicit value objects and collections over anonymous arrays or dictionaries for moves, mappings, and file context.
- Avoid duct-tape fixes. If a change needs a special case, check whether a missing abstraction is the real problem.

## Engineering standards

- Start from syntax and project configuration, not string similarity.
- Keep planning, scanning, rule execution, replacement validation, application, reporting, and validation orchestration separated.
- Rules should have explicit inputs and outputs. They should not read CLI flags, mutate package globals, or inspect Git state.
- Prefer small reusable helpers for language-neutral mechanics such as casing, byte-range handling, and replacement conversion.
- Keep language semantics in the owning adapter. A helper that knows about PHP namespaces, Python modules, or Go packages is not shared core.
- When two adapters need the same neutral helper, extract it once instead of copying almost-identical functions.
- Do not extract one-off helpers just to make code look abstract. Shared code should remove real duplication or clarify a boundary.
- Preserve original formatting by emitting byte-range replacements into the original source instead of regenerating whole files.
- Add regression tests for real bug classes. A fix without a focused test is usually not finished.
- If a file grows into mixed responsibilities, split by behaviour before adding more cases.

## CLI assumptions

- Use the explicit namespaced form: `refactorlah move ...`
- `move` is mandatory. Do not reintroduce shorthand top-level path invocation.
- Apply is the default mode. Use `--dry` for preview behaviour.
- Batch input uses `--use-list` and `--use-file`.

## Git

- Use atomic commits.
- Use conventional commits.
- Keep each commit scoped to one logical change.
- Do not bundle refactors, fixes, docs, and test changes together unless they are inseparable.
- If the user asks for amended billing, keep the result cleaner than you found it.

Examples:

- `feat: add batch move input support`
- `fix: remove redundant imports after namespace moves`
- `docs: prepare release-facing readme`
- `refactor: simplify text report layout`

## Verification

- Run `bin/test.sh` before finishing a change.
- `bin/build.sh` and `bin/install.sh` already run tests as part of normal verification.
- If you change release-facing docs or scripts, check the actual command flow they describe.

## Documentation

- Use British English in repository docs.
- Keep README release-facing and concise.
- Do not hard-code personal filesystem paths in docs or user-facing script output.
