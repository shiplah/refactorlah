# AGENTS

Use this file for repository-specific rules that are easy to miss from the code alone.

## Architecture

- Core CLI responsibilities live outside adapters: argument parsing, root/path handling, move planning, Git/filesystem moves, replacement validation/application, reporting, and validation orchestration.
- Native adapter code lives under `internal/adapters/<adapter>`.
- Adapters propose semantic results only. They must not move files, write files, inspect Git state, run validation, or print output.
- Shared semantic result types live under `internal/adapters/contract`.
- Built-in adapter registration lives under `internal/adapters/registry`.
- Parser infrastructure lives under `internal/parsing`; adapter rules do not belong there.
- Do not reintroduce external adapter executables, runtime package adapters, or process invocation fallbacks.

## Adapter Work

- Add new rewrite behaviour as a focused rule with focused tests.
- Preserve the existing reference style per occurrence where deterministic:
  - imported short name stays short
  - explicit fully-qualified name stays fully-qualified
- Keep language semantics in the owning adapter. Shared helpers should stay language-neutral.
- If a file grows into mixed responsibilities, split by behaviour before adding more cases.
- Add regression tests for real bug classes before or alongside the fix.
- Read `internal/adapters/README.md` before changing adapter code.

## Git

- Use atomic commits.
- Use conventional commits.
- Keep each commit scoped to one logical change.
- Do not bundle refactors, fixes, docs, and tests unless they are inseparable.
- If the user asks for amended history, keep the result cleaner than you found it.

## Verification

- Run `bin/test.sh` before finishing a change.
- `bin/build.sh` and `bin/install.sh` run tests unless explicitly told not to.
- If you change release-facing docs or scripts, check the command flow they describe.

## Documentation

- Use British English in repository docs.
- Keep README release-facing and concise.
- Do not hard-code personal filesystem paths in docs or user-facing script output.
