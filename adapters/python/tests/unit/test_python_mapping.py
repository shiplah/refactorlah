from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from src.protocol.request import Move
from src.project.source_roots import SourceRootResolver
from src.python.module_mapping import ModuleMapper


class PythonMappingTest(unittest.TestCase):
    def test_module_mapper_derives_src_layout_module_move(self) -> None:
        moves = (Move(old_path="src/app/services/billing.py", new_path="src/app/domain/billing.py", tracked=True),)

        mappings, warnings = ModuleMapper(("src",)).derive(moves)

        self.assertEqual((), warnings)
        self.assertEqual("app.services.billing", mappings[0].old_module)
        self.assertEqual("app.domain.billing", mappings[0].new_module)
        self.assertEqual("billing", mappings[0].old_leaf)

    def test_source_root_resolver_detects_src_layout(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "src" / "app").mkdir(parents=True)
            (root / "src" / "app" / "__init__.py").write_text("")

            roots = SourceRootResolver(root).resolve(
                (Move(old_path="src/app/old.py", new_path="src/app/new.py", tracked=True),)
            )

        self.assertIn("src", roots)

    def test_module_mapper_warns_for_unknown_source_root(self) -> None:
        mappings, warnings = ModuleMapper(("src",)).derive(
            (Move(old_path="tools/old.py", new_path="tools/new.py", tracked=False),)
        )

        self.assertEqual((), mappings)
        self.assertEqual("tools/old.py", warnings[0].file)


if __name__ == "__main__":
    unittest.main()
