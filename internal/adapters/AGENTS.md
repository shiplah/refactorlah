# AGENTS

Read `internal/adapters/README.md` before changing adapter code.

- Keep adapters as native Go code compiled into the main binary.
- Do not add external adapter executables, manifests, Composer packages, Python packages, or process invocation fallbacks.
- Adapter code must propose results only; it must not move files, write files, inspect Git state, run validation, or print output.
- Keep parser plumbing in `internal/parsing`, not in adapter rule packages.
- Keep rules narrow and independently tested beside the adapter they belong to.
- Keep shared helper code language-neutral; move semantic helpers back into the relevant adapter.
- Update `.docs/features.md` or `.docs/backlog.md` when support changes.
- Cover real-world bug reports with rule-level, adapter-level, or fixture-level regression tests.
