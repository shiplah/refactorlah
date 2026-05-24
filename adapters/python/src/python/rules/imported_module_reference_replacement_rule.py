from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class ImportedModuleReferenceReplacementRule:
    def collect(self, context: PythonFileContext, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        content = context.content
        replacements: list[Replacement] = []
        for mapping in mappings:
            if mapping.old_leaf == mapping.new_leaf:
                continue
            old_parent = _parent(mapping.old_module)
            if not old_parent or not _imports_leaf(context, old_parent, mapping.old_leaf):
                continue

            pattern = re.compile(rf"(?<![\w.]){re.escape(mapping.old_leaf)}(?=\.)")
            for match in pattern.finditer(content):
                start, end = byte_slice(content, match.start(), match.end())
                replacements.append(
                    Replacement(
                        file=context.file,
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


def _imports_leaf(context: PythonFileContext, parent: str, leaf: str) -> bool:
    absolute_pattern = re.compile(
        rf"^[ \t]*from[ \t]+{re.escape(parent)}[ \t]+import[ \t]+(?P<imports>[^\n#]*)",
        re.MULTILINE,
    )
    for match in absolute_pattern.finditer(context.content):
        if _imports_visible_leaf(match.group("imports"), leaf):
            return True

    relative_pattern = re.compile(
        r"^[ \t]*from[ \t]+(\.+)([A-Za-z_][\w.]*)?[ \t]+import[ \t]+(?P<imports>[^\n#]*)",
        re.MULTILINE,
    )
    for match in relative_pattern.finditer(context.content):
        resolved = _resolve_relative_module(context.package, len(match.group(1)), match.group(2) or "")
        if resolved == parent and _imports_visible_leaf(match.group("imports"), leaf):
            return True

    return False


def _imports_visible_leaf(import_clause: str, leaf: str) -> bool:
    for item in import_clause.split(","):
        item = item.strip()
        if not item:
            continue
        parts = re.split(r"\s+as\s+", item, maxsplit=1)
        if len(parts) > 1:
            continue
        if parts[0].strip() == leaf:
            return True
    return False


def _resolve_relative_module(package: str, level: int, module_tail: str) -> str | None:
    if not package:
        return None
    parts = package.split(".")
    up = level - 1
    if up > len(parts):
        return None
    base_parts = parts[: len(parts) - up] if up else parts
    if module_tail:
        base_parts = [*base_parts, *module_tail.split(".")]
    return ".".join(part for part in base_parts if part)
