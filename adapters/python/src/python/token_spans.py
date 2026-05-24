from __future__ import annotations

import io
import tokenize
from dataclasses import dataclass

from src.python.offsets import byte_slice


@dataclass(frozen=True)
class TokenSpanFilter:
    excluded_spans: tuple[tuple[int, int], ...]

    @classmethod
    def for_python_source(cls, content: str) -> TokenSpanFilter:
        line_offsets = _line_offsets(content)
        excluded: list[tuple[int, int]] = []

        for token in tokenize.generate_tokens(io.StringIO(content).readline):
            if token.type not in {tokenize.COMMENT, tokenize.STRING}:
                continue
            start = line_offsets[token.start[0] - 1] + token.start[1]
            end = line_offsets[token.end[0] - 1] + token.end[1]
            excluded.append(byte_slice(content, start, end))

        return cls(tuple(excluded))

    def allows(self, start: int, end: int) -> bool:
        return not any(start < excluded_end and end > excluded_start for excluded_start, excluded_end in self.excluded_spans)


def _line_offsets(content: str) -> list[int]:
    offsets = [0]
    for index, char in enumerate(content):
        if char == "\n":
            offsets.append(index + 1)
    return offsets
