# AGENTS

Read `internal/adapters/README.md` before changing adapter code.

- Keep adapters native Go code compiled into the main binary.
- Do not add external adapter executables, manifests, Composer packages, Python packages, or process invocation fallbacks.
- Keep rules narrow and independently tested beside the adapter they belong to.
- Preserve existing reference style per occurrence unless a deterministic clean-up rule proves a safer edit.
- Put parser plumbing in `internal/parsing`, not in adapter rule packages.
- Keep shared helper code small; if it starts knowing language semantics, move it back into the relevant adapter.
