from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Sequence, TextIO

from src.project.scan_policy import ScanPolicy
from src.project.source_roots import SourceRootResolver
from src.protocol.request import AdapterRequest
from src.protocol.response import AdapterResponse
from src.python.module_mapping import ModuleMapper
from src.python.reference_scanner import PythonReferenceScanner


def run(argv: Sequence[str], stdin: TextIO, stdout: TextIO, stderr: TextIO) -> int:
    if len(argv) < 2 or argv[1] != "analyze":
        stderr.write("usage: refactorlah-python analyze\n")
        return 2

    try:
        payload = json.loads(stdin.read())
        if not isinstance(payload, dict):
            raise ValueError("adapter request must decode to an object")

        request = AdapterRequest.from_payload(payload)
        project_root = Path.cwd()
        scan_policy = ScanPolicy(include=request.options.scan_include, exclude=request.options.scan_exclude)

        symbol_mappings = ()
        replacements = ()
        warnings = ()
        if request.options.include_python:
            source_roots = SourceRootResolver(project_root).resolve(request.moves)
            module_mappings, mapping_warnings = ModuleMapper(source_roots).derive(request.moves)
            replacements, scan_warnings = PythonReferenceScanner(project_root, scan_policy).scan(module_mappings)
            symbol_mappings = tuple(mapping.to_symbol_mapping() for mapping in module_mappings)
            warnings = (*mapping_warnings, *scan_warnings)

        response = AdapterResponse(
            symbol_mappings=symbol_mappings,
            replacements=replacements,
            warnings=warnings,
        )
        stdout.write(json.dumps(response.to_payload(), indent=2))
        return 0
    except Exception as exc:
        stderr.write(f"{exc}\n")
        response = AdapterResponse(errors=(str(exc),))
        stdout.write(json.dumps(response.to_payload(), indent=2))
        return 1


def main(argv: Sequence[str] | None = None) -> int:
    return run(sys.argv if argv is None else argv, sys.stdin, sys.stdout, sys.stderr)
