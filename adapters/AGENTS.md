# Adapter Agents

Read [README.md](./README.md) in this directory first.

Additional guidance for agents working in adapters:

- Preserve feature parity language with the core when the concept is shared.
- Reuse existing terminology such as `move`, wildcard support, dry mode, warnings, and deterministic rewrites.
- Prefer adding or improving extractors when multiple rules need the same file context.
- Prefer file-level coordination when multiple independent replacements would otherwise compete over imports or name resolution.
- Preserve the value-object style used by the Python adapter; do not regress new adapter code back to loose dictionary/array shapes.
- Do not let adapter internals leak into user-facing behaviour or wording unless the adapter is exposing a genuinely adapter-specific capability.

For repository-wide workflow, tests, and release-facing documentation expectations, see the `Contributing` section in the main [README](../README.md).
