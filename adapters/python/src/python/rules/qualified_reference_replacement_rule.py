from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class QualifiedReferenceReplacementRule:
    def collect(self, context: PythonFileContext, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        content = context.content
        replacements: list[Replacement] = []
        for mapping in mappings:
            pattern = re.compile(rf"(?<![\w.]){re.escape(mapping.old_module)}(?!\w)")
            for match in pattern.finditer(content):
                start, end = byte_slice(content, match.start(), match.end())
                replacements.append(
                    Replacement(
                        file=context.file,
                        start=start,
                        end=end,
                        replacement=mapping.new_module,
                        reason="python-qualified-module",
                        rule=self.__class__.__name__,
                    )
                )
        return tuple(replacements)
