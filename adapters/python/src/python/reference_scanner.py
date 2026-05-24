from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from src.files.file_collector import FileCollector
from src.project.scan_policy import ScanPolicy
from src.protocol.response import Replacement, Warning
from src.python.module_mapping import ModuleMapping
from src.python.rules.imported_module_reference_replacement_rule import ImportedModuleReferenceReplacementRule
from src.python.rules.import_replacement_rule import ImportReplacementRule
from src.python.rules.qualified_reference_replacement_rule import QualifiedReferenceReplacementRule
from src.python.rules.rule import ReplacementRule
from src.python.token_spans import TokenSpanFilter


@dataclass(frozen=True)
class PythonReferenceScanner:
    project_root: Path
    scan_policy: ScanPolicy
    rules: tuple[ReplacementRule, ...] = (
        ImportReplacementRule(),
        ImportedModuleReferenceReplacementRule(),
        QualifiedReferenceReplacementRule(),
    )

    def scan(self, mappings: tuple[ModuleMapping, ...]) -> tuple[tuple[Replacement, ...], tuple[Warning, ...]]:
        if not mappings:
            return (), ()

        files = self.scan_policy.filter(FileCollector(self.project_root).collect(("py",)))
        candidates = tuple(file for file in files if self._is_candidate(file, mappings))
        replacements: list[Replacement] = []
        warnings: list[Warning] = []

        for file in candidates:
            content = (self.project_root / file).read_text()
            token_filter = TokenSpanFilter.for_python_source(content)
            if "importlib.import_module" in content or "__import__(" in content:
                warnings.append(Warning(message="Dynamic Python import detected; not changed.", file=file))
            for rule in self.rules:
                replacements.extend(
                    replacement
                    for replacement in rule.collect(file, content, mappings)
                    if token_filter.allows(replacement.start, replacement.end)
                )

        return tuple(_deduplicate(replacements)), tuple(warnings)

    def _is_candidate(self, file: str, mappings: tuple[ModuleMapping, ...]) -> bool:
        content = (self.project_root / file).read_text()
        return any(
            mapping.old_path == file
            or mapping.old_module in content
            or mapping.old_leaf in content
            for mapping in mappings
        )


def _deduplicate(replacements: list[Replacement]) -> list[Replacement]:
    seen: set[tuple[str, int, int, str]] = set()
    unique: list[Replacement] = []
    for replacement in replacements:
        key = (replacement.file, replacement.start, replacement.end, replacement.replacement)
        if key in seen:
            continue
        seen.add(key)
        unique.append(replacement)
    return unique
