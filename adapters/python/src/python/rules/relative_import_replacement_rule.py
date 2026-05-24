from __future__ import annotations

import re
from dataclasses import dataclass

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class RelativeImportReplacementRule:
    def collect(self, context: PythonFileContext, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        replacements: list[Replacement] = []
        pattern = re.compile(r"(^[ \t]*from[ \t]+)(\.+)([A-Za-z_][\w.]*)?([ \t]+import[ \t]+)([^\n#]+)", re.MULTILINE)

        for match in pattern.finditer(context.content):
            module_tail = match.group(3) or ""
            resolved_module = _resolve_relative_module(context.package, len(match.group(2)), module_tail)
            if resolved_module is None:
                continue

            for mapping in mappings:
                if module_tail and resolved_module == mapping.old_module:
                    replacements.append(
                        _replacement(
                            context,
                            match.start(2),
                            match.end(3),
                            mapping.new_module,
                            "python-relative-from-import",
                            self.__class__.__name__,
                        )
                    )
                    continue

                if module_tail:
                    continue

                old_parent = _parent(mapping.old_module)
                new_parent = _parent(mapping.new_module)
                if resolved_module != old_parent or not _imports_name(match.group(5), mapping.old_leaf):
                    continue
                if not new_parent:
                    continue

                replacements.append(
                    _replacement(
                        context,
                        match.start(2),
                        match.end(2),
                        new_parent,
                        "python-relative-from-import",
                        self.__class__.__name__,
                    )
                )

                if mapping.old_leaf != mapping.new_leaf:
                    leaf_start = _imported_name_offset(match.group(5), match.start(5), mapping.old_leaf)
                    if leaf_start is not None:
                        replacements.append(
                            _replacement(
                                context,
                                leaf_start,
                                leaf_start + len(mapping.old_leaf),
                                mapping.new_leaf,
                                "python-relative-from-import-name",
                                self.__class__.__name__,
                            )
                        )

        return tuple(replacements)


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


def _parent(module: str) -> str:
    return module.rsplit(".", 1)[0] if "." in module else ""


def _imports_name(import_clause: str, name: str) -> bool:
    return _imported_name_offset(import_clause, 0, name) is not None


def _imported_name_offset(import_clause: str, clause_start: int, name: str) -> int | None:
    pattern = re.compile(rf"(?<!\w){re.escape(name)}(?!\w)")
    match = pattern.search(import_clause)
    if match is None:
        return None
    return clause_start + match.start()


def _replacement(context: PythonFileContext, start: int, end: int, replacement: str, reason: str, rule: str) -> Replacement:
    byte_start, byte_end = byte_slice(context.content, start, end)
    return Replacement(
        file=context.file,
        start=byte_start,
        end=byte_end,
        replacement=replacement,
        reason=reason,
        rule=rule,
    )
