from __future__ import annotations

from typing import Protocol

from src.protocol.response import Replacement
from src.python.module_mapping import ModuleMapping


class ReplacementRule(Protocol):
    def collect(self, file: str, content: str, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        ...
