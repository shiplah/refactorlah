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

Safety behaviour:

- comments are not rewritten
- arbitrary strings are not rewritten
- string annotations and docstrings are not rewritten yet
- dynamic imports are reported as warnings
- files excluded by `.refactorlah.json` are not semantically scanned

## Design Notes

The Python adapter intentionally starts with value objects and explicit file context:

- protocol payloads become dataclasses
- moves become `Move` objects
- module mappings become `ModuleMapping` objects
- rules receive `PythonFileContext`

Keep that style. Do not introduce loose dictionaries for data that has stable shape.

## Future Work

Likely next additions:

- string annotation rewrites where resolvable
- richer relative import resolution
- Jinja/Django template support under framework-specific namespaces
- Python validation hooks such as `pytest`, `mypy`, or `ruff`
