from __future__ import annotations

from dataclasses import dataclass
from fnmatch import fnmatchcase


@dataclass(frozen=True)
class ScanPolicy:
    include: tuple[str, ...]
    exclude: tuple[str, ...]

    def allows(self, path: str) -> bool:
        for pattern in self.include:
            if fnmatchcase(path, pattern):
                return True

        for pattern in self.exclude:
            if fnmatchcase(path, pattern):
                return False

        return True

    def filter(self, paths: tuple[str, ...]) -> tuple[str, ...]:
        return tuple(path for path in paths if self.allows(path))
