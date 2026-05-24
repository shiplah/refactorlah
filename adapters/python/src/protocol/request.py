from __future__ import annotations

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class Move:
    old_path: str
    new_path: str
    tracked: bool


@dataclass(frozen=True)
class RequestOptions:
    include_python: bool
    scan_include: tuple[str, ...]
    scan_exclude: tuple[str, ...]


@dataclass(frozen=True)
class AdapterRequest:
    old_path: str
    new_path: str
    dry_run: bool
    moves: tuple[Move, ...]
    options: RequestOptions

    @classmethod
    def from_payload(cls, payload: dict[str, Any]) -> AdapterRequest:
        if _int(payload.get("protocolVersion")) != 1:
            raise ValueError("adapter request must use protocolVersion 1")
        if _str(payload.get("projectRoot")) != ".":
            raise ValueError('adapter request must use projectRoot "."')
        if not isinstance(payload.get("dryRun"), bool):
            raise ValueError("adapter request must include dryRun")

        old_path = _str(payload.get("oldPath"))
        new_path = _str(payload.get("newPath"))
        if not old_path or not new_path:
            raise ValueError("adapter request must include oldPath and newPath")

        moves = tuple(_moves(payload.get("moves")))
        if not moves:
            raise ValueError("adapter request must include at least one move")

        return cls(
            old_path=old_path,
            new_path=new_path,
            dry_run=bool(payload["dryRun"]),
            moves=moves,
            options=_options(payload.get("options")),
        )


def _moves(value: object) -> list[Move]:
    if not isinstance(value, list):
        return []

    moves: list[Move] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        old_path = _str(item.get("oldPath"))
        new_path = _str(item.get("newPath"))
        if not old_path or not new_path:
            continue
        moves.append(Move(old_path=old_path, new_path=new_path, tracked=bool(item.get("tracked"))))
    return moves


def _options(value: object) -> RequestOptions:
    if not isinstance(value, dict):
        return RequestOptions(include_python=False, scan_include=(), scan_exclude=())

    return RequestOptions(
        include_python=bool(value.get("includePython")),
        scan_include=tuple(_string_list(value.get("scanInclude"))),
        scan_exclude=tuple(_string_list(value.get("scanExclude"))),
    )


def _string_list(value: object) -> list[str]:
    if not isinstance(value, list):
        return []
    return [item for item in value if isinstance(item, str) and item]


def _int(value: object) -> int:
    return value if isinstance(value, int) else 0


def _str(value: object) -> str:
    return value if isinstance(value, str) else ""
