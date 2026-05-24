from __future__ import annotations

from dataclasses import dataclass
from pathlib import PurePosixPath

from src.protocol.request import Move
from src.protocol.response import SymbolMapping, Warning


@dataclass(frozen=True)
class ModuleMapping:
    old_path: str
    new_path: str
    old_module: str
    new_module: str
    old_leaf: str
    new_leaf: str

    def to_symbol_mapping(self) -> SymbolMapping:
        return SymbolMapping(
            kind="module",
            old_path=self.old_path,
            new_path=self.new_path,
            old_symbol=self.old_module,
            new_symbol=self.new_module,
            old_namespace=self._namespace(self.old_module),
            new_namespace=self._namespace(self.new_module),
            short_name=self.old_leaf,
        )

    @staticmethod
    def _namespace(module: str) -> str:
        return module.rsplit(".", 1)[0] if "." in module else ""


@dataclass(frozen=True)
class ModuleMapper:
    source_roots: tuple[str, ...]

    def derive(self, moves: tuple[Move, ...]) -> tuple[tuple[ModuleMapping, ...], tuple[Warning, ...]]:
        mappings: list[ModuleMapping] = []
        warnings: list[Warning] = []

        for move in moves:
            if not move.old_path.endswith(".py") or not move.new_path.endswith(".py"):
                continue

            old_module = self._module_for_path(move.old_path)
            new_module = self._module_for_path(move.new_path)
            if old_module is None or new_module is None:
                warnings.append(Warning(message="Python file is outside known source roots; semantic rewrites skipped", file=move.old_path))
                continue
            if old_module == new_module:
                continue

            mappings.append(
                ModuleMapping(
                    old_path=move.old_path,
                    new_path=move.new_path,
                    old_module=old_module,
                    new_module=new_module,
                    old_leaf=old_module.rsplit(".", 1)[-1],
                    new_leaf=new_module.rsplit(".", 1)[-1],
                )
            )

        return tuple(mappings), tuple(warnings)

    def _module_for_path(self, path: str) -> str | None:
        clean = path.strip("/")
        for root in self.source_roots:
            prefix = "" if root == "." else root.rstrip("/") + "/"
            if prefix and not clean.startswith(prefix):
                continue
            relative = clean[len(prefix) :] if prefix else clean
            if not relative.endswith(".py"):
                continue
            module_path = PurePosixPath(relative[:-3])
            parts = [part for part in module_path.parts if part and part != "__init__"]
            if not parts:
                continue
            return ".".join(parts)
        return None
