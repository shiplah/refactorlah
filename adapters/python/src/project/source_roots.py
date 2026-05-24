from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from src.protocol.request import Move


@dataclass(frozen=True)
class SourceRootResolver:
    project_root: Path

    def resolve(self, moves: tuple[Move, ...]) -> tuple[str, ...]:
        roots: set[str] = set()
        if (self.project_root / "src").is_dir():
            roots.add("src")
        if self._has_top_level_packages():
            roots.add(".")

        for move in moves:
            for path in (move.old_path, move.new_path):
                root = self._nearest_package_parent(path)
                if root is not None:
                    roots.add(root)

        if not roots:
            roots.add("src" if any(move.old_path.startswith("src/") or move.new_path.startswith("src/") for move in moves) else ".")

        return tuple(sorted(roots, key=lambda root: (root.count("/"), len(root)), reverse=True))

    def _has_top_level_packages(self) -> bool:
        for child in self.project_root.iterdir():
            if child.is_dir() and (child / "__init__.py").is_file():
                return True
        return False

    def _nearest_package_parent(self, path: str) -> str | None:
        current = self.project_root / Path(path).parent
        package_root = current
        saw_package = False

        while current != self.project_root and (current / "__init__.py").is_file():
            saw_package = True
            package_root = current
            current = current.parent

        if not saw_package:
            return None

        parent = package_root.parent
        if parent == self.project_root:
            return "."
        return parent.relative_to(self.project_root).as_posix()
