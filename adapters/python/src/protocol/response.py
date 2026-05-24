from __future__ import annotations

from dataclasses import asdict, dataclass
from typing import Any


@dataclass(frozen=True)
class SymbolMapping:
    kind: str
    old_path: str
    new_path: str
    old_symbol: str
    new_symbol: str
    old_namespace: str = ""
    new_namespace: str = ""
    short_name: str = ""

    def to_payload(self) -> dict[str, Any]:
        return _camel_payload(asdict(self))


@dataclass(frozen=True)
class PathMapping:
    kind: str
    old_path: str
    new_path: str
    old_reference: str
    new_reference: str

    def to_payload(self) -> dict[str, Any]:
        return _camel_payload(asdict(self))


@dataclass(frozen=True)
class Replacement:
    file: str
    start: int
    end: int
    replacement: str
    reason: str
    rule: str = ""

    def to_payload(self) -> dict[str, Any]:
        payload = asdict(self)
        if not payload["rule"]:
            del payload["rule"]
        return payload


@dataclass(frozen=True)
class Warning:
    message: str
    file: str = ""
    line: int = 0

    def to_payload(self) -> dict[str, Any]:
        payload: dict[str, Any] = {"message": self.message}
        if self.file:
            payload["file"] = self.file
        if self.line:
            payload["line"] = self.line
        return payload


@dataclass(frozen=True)
class AdapterResponse:
    symbol_mappings: tuple[SymbolMapping, ...] = ()
    path_mappings: tuple[PathMapping, ...] = ()
    replacements: tuple[Replacement, ...] = ()
    warnings: tuple[Warning, ...] = ()
    errors: tuple[str, ...] = ()

    def to_payload(self) -> dict[str, Any]:
        return {
            "protocolVersion": 1,
            "adapter": "python",
            "symbolMappings": [mapping.to_payload() for mapping in self.symbol_mappings],
            "pathMappings": [mapping.to_payload() for mapping in self.path_mappings],
            "replacements": [replacement.to_payload() for replacement in self.replacements],
            "warnings": [warning.to_payload() for warning in self.warnings],
            "errors": list(self.errors),
        }


def _camel_payload(values: dict[str, Any]) -> dict[str, Any]:
    return {
        "kind": values["kind"],
        "oldPath": values["old_path"],
        "newPath": values["new_path"],
        "oldSymbol": values["old_symbol"],
        "newSymbol": values["new_symbol"],
        "oldNamespace": values.get("old_namespace", ""),
        "newNamespace": values.get("new_namespace", ""),
        "shortName": values.get("short_name", ""),
    } if "old_symbol" in values else {
        "kind": values["kind"],
        "oldPath": values["old_path"],
        "newPath": values["new_path"],
        "oldReference": values["old_reference"],
        "newReference": values["new_reference"],
    }
