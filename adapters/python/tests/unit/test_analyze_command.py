from __future__ import annotations

import io
import json
import os
import tempfile
import unittest
from pathlib import Path

from src.analyze_command import run
from src.protocol.response import Replacement


class AnalyzeCommandTest(unittest.TestCase):
    def test_analyze_command_updates_python_imports_and_references(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            write(root / "src" / "app" / "__init__.py", "")
            write(root / "src" / "app" / "services" / "billing.py", "class InvoiceService: pass\n")
            write(root / "src" / "app" / "http" / "controller.py", "import app.services.billing\nservice = app.services.billing.InvoiceService()\n")

            decoded = run_adapter(
                root,
                {
                    "protocolVersion": 1,
                    "projectRoot": ".",
                    "oldPath": "src/app/services/billing.py",
                    "newPath": "src/app/domain/billing.py",
                    "dryRun": True,
                    "moves": [
                        {
                            "oldPath": "src/app/services/billing.py",
                            "newPath": "src/app/domain/billing.py",
                            "tracked": True,
                        }
                    ],
                    "options": {"includePython": True},
                },
            )

            content = (root / "src" / "app" / "http" / "controller.py").read_text()
            updated = apply_replacements(content, replacements_for(decoded, "src/app/http/controller.py"))

        self.assertEqual("import app.domain.billing\nservice = app.domain.billing.InvoiceService()\n", updated)
        self.assertEqual("app.services.billing", decoded["symbolMappings"][0]["oldSymbol"])
        self.assertEqual("app.domain.billing", decoded["symbolMappings"][0]["newSymbol"])

    def test_analyze_command_honours_scan_excludes(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            write(root / "src" / "app" / "__init__.py", "")
            write(root / "src" / "app" / "services" / "billing.py", "class InvoiceService: pass\n")
            write(root / "src" / "app" / "generated" / "fixture.py", "import app.services.billing\n")

            decoded = run_adapter(
                root,
                {
                    "protocolVersion": 1,
                    "projectRoot": ".",
                    "oldPath": "src/app/services/billing.py",
                    "newPath": "src/app/domain/billing.py",
                    "dryRun": True,
                    "moves": [
                        {
                            "oldPath": "src/app/services/billing.py",
                            "newPath": "src/app/domain/billing.py",
                            "tracked": True,
                        }
                    ],
                    "options": {"includePython": True, "scanExclude": ["src/app/generated/**"]},
                },
            )

        self.assertEqual([], decoded["replacements"])

    def test_analyze_command_warns_on_dynamic_imports(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            write(root / "src" / "app" / "__init__.py", "")
            write(root / "src" / "app" / "services" / "billing.py", "class InvoiceService: pass\n")
            write(root / "src" / "app" / "consumer.py", "importlib.import_module(name)\n# app.services.billing\n")

            decoded = run_adapter(
                root,
                {
                    "protocolVersion": 1,
                    "projectRoot": ".",
                    "oldPath": "src/app/services/billing.py",
                    "newPath": "src/app/domain/billing.py",
                    "dryRun": True,
                    "moves": [
                        {
                            "oldPath": "src/app/services/billing.py",
                            "newPath": "src/app/domain/billing.py",
                            "tracked": True,
                        }
                    ],
                    "options": {"includePython": True},
                },
            )

        self.assertEqual("Dynamic Python import detected; not changed.", decoded["warnings"][0]["message"])

    def test_analyze_command_does_not_rewrite_comments_or_strings(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            write(root / "src" / "app" / "__init__.py", "")
            write(root / "src" / "app" / "services" / "billing.py", "class InvoiceService: pass\n")
            write(
                root / "src" / "app" / "consumer.py",
                "# app.services.billing\nvalue = 'app.services.billing'\nimport app.services.billing\n",
            )

            decoded = run_adapter(
                root,
                {
                    "protocolVersion": 1,
                    "projectRoot": ".",
                    "oldPath": "src/app/services/billing.py",
                    "newPath": "src/app/domain/billing.py",
                    "dryRun": True,
                    "moves": [
                        {
                            "oldPath": "src/app/services/billing.py",
                            "newPath": "src/app/domain/billing.py",
                            "tracked": True,
                        }
                    ],
                    "options": {"includePython": True},
                },
            )

            content = (root / "src" / "app" / "consumer.py").read_text()
            updated = apply_replacements(content, replacements_for(decoded, "src/app/consumer.py"))

        self.assertEqual("# app.services.billing\nvalue = 'app.services.billing'\nimport app.domain.billing\n", updated)


def run_adapter(root: Path, payload: dict[str, object]) -> dict[str, object]:
    previous = Path.cwd()
    try:
        os.chdir(root)
        stdout = io.StringIO()
        stderr = io.StringIO()
        code = run(["refactorlah-python", "analyze"], io.StringIO(json.dumps(payload)), stdout, stderr)
    finally:
        os.chdir(previous)

    if code != 0:
        raise AssertionError(stderr.getvalue())
    return json.loads(stdout.getvalue())


def write(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def replacements_for(decoded: dict[str, object], file: str) -> tuple[Replacement, ...]:
    replacements = []
    for item in decoded["replacements"]:
        assert isinstance(item, dict)
        if item["file"] != file:
            continue
        replacements.append(
            Replacement(
                file=str(item["file"]),
                start=int(item["start"]),
                end=int(item["end"]),
                replacement=str(item["replacement"]),
                reason=str(item["reason"]),
            )
        )
    return tuple(replacements)


def apply_replacements(content: str, replacements: tuple[Replacement, ...]) -> str:
    result = content.encode()
    for replacement in sorted(replacements, key=lambda item: item.start, reverse=True):
        result = result[: replacement.start] + replacement.replacement.encode() + result[replacement.end :]
    return result.decode()


if __name__ == "__main__":
    unittest.main()
