from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from src.files.file_collector import FileCollector


class FileCollectorTest(unittest.TestCase):
    def test_collector_prunes_ignored_directories(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            write(root / "src" / "app" / "service.py", "")
            write(root / "node_modules" / "package" / "ignored.py", "")
            write(root / "src" / "app" / "__pycache__" / "ignored.py", "")
            write(root / ".venv" / "lib" / "ignored.py", "")

            files = FileCollector(root).collect(("py",))

        self.assertEqual(("src/app/service.py",), files)


def write(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


if __name__ == "__main__":
    unittest.main()
