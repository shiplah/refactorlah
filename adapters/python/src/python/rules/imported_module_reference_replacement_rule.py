from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class ImportedModuleReferenceReplacementRule:
    def collect(self, file: str, content: str, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        replacements: list[Replacement] = []
        for mapping in mappings:
            if mapping.old_leaf == mapping.new_leaf:
                continue
            old_parent = _parent(mapping.old_module)
            if not old_parent or not _imports_leaf(content, old_parent, mapping.old_leaf):
                continue

            pattern = re.compile(rf"(?<![\w.]){re.escape(mapping.old_leaf)}(?=\.)")
            for match in pattern.finditer(content):
                start, end = byte_slice(content, match.start(), match.end())
                replacements.append(
                    Replacement(
                        file=file,
                        start=start,
                        end=end,
                        replacement=mapping.new_leaf,
                        reason="python-imported-module-reference",
                        rule=self.__class__.__name__,
                    )
                )
        return tuple(replacements)


def _parent(module: str) -> str:
    return module.rsplit(".", 1)[0] if "." in module else ""


def _imports_leaf(content: str, parent: str, leaf: str) -> bool:
    pattern = re.compile(
        rf"^[ \t]*from[ \t]+{re.escape(parent)}[ \t]+import[ \t]+[^\n#]*\b{re.escape(leaf)}\b",
        re.MULTILINE,
    )
    return bool(pattern.search(content))
