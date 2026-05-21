# AGENTS

## Commit style

- Use atomic commits.
- Use conventional commits for commit messages.
- Keep each commit scoped to one logical change.

Examples:

- `feat: add batch move input support`
- `fix: preserve namespace rewrites for moved tests`
- `docs: clarify move command usage`
- `refactor: simplify root command routing`

## Command usage

- Use the explicit namespaced CLI form: `refactorlah move ...`
- Do not rely on shorthand top-level path invocation.

## Verification

- Run `bin/test.sh` before finishing a change.
- Expect build/install flows to exercise tests as part of normal verification.
