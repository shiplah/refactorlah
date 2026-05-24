from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class QualifiedReferenceReplacementRule:
    def collect(self, file: str, content: str, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        replacements: list[Replacement] = []
        for mapping in mappings:
            pattern = re.compile(rf"(?<![\w.]){re.escape(mapping.old_module)}(?!\w)")
            for match in pattern.finditer(content):
                start, end = byte_slice(content, match.start(), match.end())
                replacements.append(
                    Replacement(
                        file=file,
                        start=start,
                        end=end,
                        replacement=mapping.new_module,
                        reason="python-qualified-module",
                        rule=self.__class__.__name__,
                    )
                )
        return tuple(replacements)
