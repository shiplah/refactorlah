from __future__ import annotations


def byte_offset(content: str, char_offset: int) -> int:
    return len(content[:char_offset].encode())


def byte_slice(content: str, start: int, end: int) -> tuple[int, int]:
    return byte_offset(content, start), byte_offset(content, end)
