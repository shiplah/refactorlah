from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path


IGNORED_PREFIXES = (
    ".git/",
    ".venv/",
    "__pycache__/",
    "build/",
    "coverage/",
    "dist/",
    "node_modules/",
    "var/",
    "vendor/",
)


@dataclass(frozen=True)
class FileCollector:
    project_root: Path

    def collect(self, extensions: tuple[str, ...]) -> tuple[str, ...]:
        allowed = {extension.lower().lstrip(".") for extension in extensions}
        paths: list[str] = []

        for path in self.project_root.rglob("*"):
            if not path.is_file():
                continue
            relative = path.relative_to(self.project_root).as_posix()
            if self._is_ignored(relative):
                continue
            if path.suffix.lower().lstrip(".") not in allowed:
                continue
            paths.append(relative)

        return tuple(sorted(paths))

    def _is_ignored(self, path: str) -> bool:
        return any(path == prefix[:-1] or path.startswith(prefix) for prefix in IGNORED_PREFIXES)
