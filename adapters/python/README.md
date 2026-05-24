# Python Adapter

`refactorlah-python` analyses Python projects for deterministic module moves.

It follows the shared adapter protocol:

- read JSON from stdin
- write JSON to stdout
- never modify files directly
- return byte-offset replacement proposals and warnings

## Current Scope

The adapter currently supports `.py` file moves where old and new paths can be mapped to importable module names.

Supported rewrites:

- `import old.module`
- `from old.module import Name`
- `from old.parent import module`
- safe relative imports such as `from .module import Name`
- safe relative parent imports such as `from . import module`
- visible imported module references when a module basename changes
- fully qualified module references such as `old.module.Name`
- exact module references inside string annotations
- exact dotted module references in `.toml`, `.ini`, `.cfg`, `.yaml`, and `.yml` config files

Safety behaviour:

- comments are not rewritten
- arbitrary strings are not rewritten
- docstrings are not rewritten
- dynamic imports are reported as warnings
- files excluded by `.refactorlah.json` are not semantically scanned

Validation support is orchestrated by the Go core, not this adapter:

- configured `ruff` and `mypy` run after apply when the executable is available
- configured `pytest` runs only when `--run-tests` is passed and the executable is available

## Design Notes

The Python adapter intentionally starts with value objects and explicit file context:

- protocol payloads become dataclasses
- moves become `Move` objects
- module mappings become `ModuleMapping` objects
- rules receive `PythonFileContext`

Keep that style. Do not introduce loose dictionaries for data that has stable shape.

## Future Work

Likely next additions:

- Jinja/Django template support under framework-specific namespaces
- richer package and project-root detection for unusual layouts
- framework-specific string/config rules where they can be proven safely
