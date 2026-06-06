# AGENTS

Read `internal/adapters/README.md` before changing adapter code.

- Keep adapters native Go code compiled into the main binary.
- Do not add external adapter executables, manifests, Composer packages, Python packages, or process invocation fallbacks.
- Keep rules narrow and independently tested beside the adapter they belong to.
- Preserve existing reference style per occurrence unless a deterministic clean-up rule proves a safer edit.
- Put parser plumbing in `internal/parsing`, not in adapter rule packages.
- Keep shared helper code small; if it starts knowing language semantics, move it back into the relevant adapter.
- Maintain feature parity where concepts overlap across adapters; update the support matrix when parity changes.
- Cover real-world bug reports with rule-level, adapter-level, or fixture-level regression tests.
- Prefer explicit value objects and typed inputs over maps, anonymous structs, or loosely shaped data.
- Framework-specific rules belong under the owning language and framework namespace, such as `php/symfony/twig`.
- If a rewrite is not provable from project configuration and syntax context, skip it and report a warning.
