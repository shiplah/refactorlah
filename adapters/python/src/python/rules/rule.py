from __future__ import annotations

from typing import Protocol

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping


class ReplacementRule(Protocol):
    def collect(self, context: PythonFileContext, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        ...
