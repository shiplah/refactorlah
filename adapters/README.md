# Adapters

Adapters extend `refactorlah` with language- or framework-specific analysis.

Current adapters:

- PHP, including Composer/PSR-4 and Symfony/Twig support.
- Python, including module moves and deterministic import rewrites.

The core CLI owns:

- argument parsing
- move planning
- adapter discovery
- filesystem and Git moves
- replacement validation
- writing files
- validation commands
- reporting

Adapters own:

- project inspection for their language or framework
- deterministic symbol and path mapping
- replacement proposals
- warnings for uncertain cases

Adapters must not write files. They propose replacements and warnings; the core decides whether and how those replacements are applied.

## What we expect from adapters

- Match core CLI behaviour where the concept exists.
- Aim for feature parity across adapters over time.
- Use the same product language as the core where the behaviour is shared.

Examples:

- If the core supports wildcard-driven move requests, adapter behaviour must still make sense for wildcard-expanded moves.
- If the core is conservative about uncertain rewrites, adapters must be conservative too.
- If the core reports warnings instead of guessing, adapters must do the same.

That does not mean every adapter must support every language feature on day one. It means shared concepts should line up, and missing capability should be explicit rather than implicit.

## Design expectations

- Keep adapter logic deterministic.
- Prefer small focused extractors and rules over large scanners.
- Prefer value objects and collections for protocol moves, mappings, file context, and rule inputs.
- Use extractors for shared file context and fact gathering.
- Use rules for narrow rewrite decisions based on those facts.
- Keep replacement generation testable in isolation.
- Keep file-level coordination explicit when syntax areas are coupled, such as imports and name resolution.

## Compatibility expectations

- Follow the JSON stdin/stdout protocol used by the core.
- Keep an `adapter.json` manifest at the adapter root with the adapter executable and external runtime requirement.
- Keep runtime execution and version-check implementation code-owned in the Go core; do not encode executable snippets in manifests.
- Treat adapter package files as build/test responsibility, not as target-project prerequisites in the manifest.
- Use project-relative slash paths in protocol messages.
- Keep CLI-facing semantics aligned with the core command surface and wording where possible.
- Do not invent adapter-specific user workflows when the core already has a shared concept.
- Honour the core-provided scan include/exclude rules for generated, fixture, and stub files where the adapter performs semantic scans.

## Adding a new adapter

When adding a new adapter:

- document its supported scope and conservative skips
- add focused unit tests for extractors and rules
- add at least one end-to-end fixture or integration path
- align naming and behaviour with existing adapters where the concepts match

## Changing an existing adapter

When changing an existing adapter:

- prefer extending shared extractor context over duplicating lookup logic
- add regression tests for the final applied file shape, not just proposed replacements
- check whether a behaviour belongs in a focused rule or in file-level coordination

For repository-wide workflow and verification expectations, see the `Contributing` section in the main [README](../README.md).
