from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_offset, byte_slice


@dataclass(frozen=True)
class ImportReplacementRule:
    def collect(self, file: str, content: str, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        replacements: list[Replacement] = []
        for mapping in mappings:
            replacements.extend(self._import_replacements(file, content, mapping))
            replacements.extend(self._from_import_replacements(file, content, mapping))
        return tuple(replacements)

    def _import_replacements(self, file: str, content: str, mapping: ModuleMapping) -> list[Replacement]:
        pattern = re.compile(rf"(^[ \t]*import[ \t]+){re.escape(mapping.old_module)}(?=([ \t]+as[ \t]+\w+)?[ \t]*(?:#.*)?$)", re.MULTILINE)
        return [
            Replacement(
                file=file,
                start=byte_offset(content, match.start(0) + len(match.group(1))),
                end=byte_offset(content, match.start(0) + len(match.group(1)) + len(mapping.old_module)),
                replacement=mapping.new_module,
                reason="python-import",
                rule=self.__class__.__name__,
            )
            for match in pattern.finditer(content)
        ]

    def _from_import_replacements(self, file: str, content: str, mapping: ModuleMapping) -> list[Replacement]:
        replacements: list[Replacement] = []

        exact_pattern = re.compile(rf"(^[ \t]*from[ \t]+){re.escape(mapping.old_module)}(?=[ \t]+import[ \t]+)", re.MULTILINE)
        for match in exact_pattern.finditer(content):
            start, end = byte_slice(
                content,
                match.start(0) + len(match.group(1)),
                match.start(0) + len(match.group(1)) + len(mapping.old_module),
            )
            replacements.append(
                Replacement(
                    file=file,
                    start=start,
                    end=end,
                    replacement=mapping.new_module,
                    reason="python-from-import",
                    rule=self.__class__.__name__,
                )
            )

        old_parent = _parent(mapping.old_module)
        new_parent = _parent(mapping.new_module)
        if old_parent and new_parent and old_parent != new_parent:
            parent_pattern = re.compile(
                rf"(^[ \t]*from[ \t]+){re.escape(old_parent)}(?=[ \t]+import[ \t]+[^\n#]*\b{re.escape(mapping.old_leaf)}\b)",
                re.MULTILINE,
            )
            for match in parent_pattern.finditer(content):
                start, end = byte_slice(
                    content,
                    match.start(0) + len(match.group(1)),
                    match.start(0) + len(match.group(1)) + len(old_parent),
                )
                replacements.append(
                    Replacement(
                        file=file,
                        start=start,
                        end=end,
                        replacement=new_parent,
                        reason="python-from-import",
                        rule=self.__class__.__name__,
                    )
                )

        if mapping.old_leaf != mapping.new_leaf:
            leaf_pattern = re.compile(
                rf"(^[ \t]*from[ \t]+{re.escape(old_parent)}[ \t]+import[ \t]+[^\n#]*?)\b{re.escape(mapping.old_leaf)}\b",
                re.MULTILINE,
            )
            for match in leaf_pattern.finditer(content):
                start, end = byte_slice(content, match.end(1), match.end(1) + len(mapping.old_leaf))
                replacements.append(
                    Replacement(
                        file=file,
                        start=start,
                        end=end,
                        replacement=mapping.new_leaf,
                        reason="python-from-import-name",
                        rule=self.__class__.__name__,
                    )
                )

        return replacements


def _parent(module: str) -> str:
    return module.rsplit(".", 1)[0] if "." in module else ""
