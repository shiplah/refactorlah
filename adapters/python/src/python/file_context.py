from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class PythonFileContext:
    file: str
    content: str
    module: str
    package: str
