from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path


IGNORED_DIRECTORY_NAMES = (
    ".git",
    ".mypy_cache",
    ".pytest_cache",
    ".ruff_cache",
    ".venv",
    "__pycache__",
    "build",
    "coverage",
    "dist",
    "node_modules",
    "var",
    "vendor",
    "venv",
)

IGNORED_PATH_PREFIXES = (
    "bootstrap/cache/",
    "storage/framework/",
)


@dataclass(frozen=True)
class FileCollector:
    project_root: Path

    def collect(self, extensions: tuple[str, ...]) -> tuple[str, ...]:
        allowed = {extension.lower().lstrip(".") for extension in extensions}
        paths: list[str] = []

        for current, directories, files in os.walk(self.project_root):
            current_path = Path(current)
            relative_directory = current_path.relative_to(self.project_root).as_posix()
            if relative_directory == ".":
                relative_directory = ""

            directories[:] = sorted(
                directory for directory in directories if not self._is_ignored_directory(relative_directory, directory)
            )

            for file in sorted(files):
                path = current_path / file
                relative = path.relative_to(self.project_root).as_posix()
                if self._is_ignored(relative):
                    continue
                if path.suffix.lower().lstrip(".") not in allowed:
                    continue
                paths.append(relative)

        return tuple(paths)

    def _is_ignored_directory(self, parent: str, directory: str) -> bool:
        if directory in IGNORED_DIRECTORY_NAMES:
            return True

        relative = f"{parent}/{directory}" if parent else directory
        return self._is_ignored(relative)

    def _is_ignored(self, path: str) -> bool:
        parts = path.split("/")
        if any(part in IGNORED_DIRECTORY_NAMES for part in parts):
            return True

        directory_path = path if path.endswith("/") else f"{path}/"
        return any(directory_path.startswith(prefix) for prefix in IGNORED_PATH_PREFIXES)
