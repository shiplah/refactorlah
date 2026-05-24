from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path

from src.files.file_collector import FileCollector
from src.project.scan_policy import ScanPolicy
from src.protocol.response import Replacement
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


CONFIG_EXTENSIONS = ("cfg", "ini", "toml", "yaml", "yml")
DOT_PATH_CHARACTER = re.compile(r"[A-Za-z0-9_.]")
MODULE_SUFFIX_CHARACTER = re.compile(r"[A-Za-z0-9_]")


@dataclass(frozen=True)
class DottedPathReferenceScanner:
    project_root: Path
    scan_policy: ScanPolicy

    def scan(self, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        if not mappings:
            return ()

        files = self.scan_policy.filter(FileCollector(self.project_root).collect(CONFIG_EXTENSIONS))
        replacements: list[Replacement] = []
        for file in files:
            content = (self.project_root / file).read_text()
            for mapping in mappings:
                if mapping.old_module not in content:
                    continue
                replacements.extend(self._replace_module_references(file, content, mapping))

        return tuple(replacements)

    def _replace_module_references(self, file: str, content: str, mapping: ModuleMapping) -> tuple[Replacement, ...]:
        replacements: list[Replacement] = []
        for match in re.finditer(re.escape(mapping.old_module), content):
            if not _is_safe_dotted_path_match(content, match.start(), match.end()):
                continue
            if _is_comment_line(content, match.start()):
                continue

            start, end = byte_slice(content, match.start(), match.end())
            replacements.append(
                Replacement(
                    file=file,
                    start=start,
                    end=end,
                    replacement=mapping.new_module,
                    reason="python-config-dotted-path",
                    rule=self.__class__.__name__,
                )
            )

        return tuple(replacements)


def _is_safe_dotted_path_match(content: str, start: int, end: int) -> bool:
    previous = content[start - 1] if start > 0 else ""
    following = content[end] if end < len(content) else ""

    if previous and DOT_PATH_CHARACTER.match(previous):
        return False
    if following and MODULE_SUFFIX_CHARACTER.match(following):
        return False
    return True


def _is_comment_line(content: str, offset: int) -> bool:
    line_start = content.rfind("\n", 0, offset) + 1
    return content[line_start:offset].lstrip().startswith(("#", ";"))

