from __future__ import annotations

import io
import json
import unittest

from src.analyze_command import run
from src.protocol.request import AdapterRequest
from src.protocol.response import AdapterResponse, Replacement, SymbolMapping, Warning


class ProtocolTest(unittest.TestCase):
    def test_request_parses_protocol_payload(self) -> None:
        request = AdapterRequest.from_payload(
            {
                "protocolVersion": 1,
                "projectRoot": ".",
                "oldPath": "src/app/old.py",
                "newPath": "src/app/new.py",
                "dryRun": True,
                "moves": [{"oldPath": "src/app/old.py", "newPath": "src/app/new.py", "tracked": True}],
                "options": {
                    "includePython": True,
                    "scanInclude": ["src/app/keep.py"],
                    "scanExclude": ["src/app/generated/**"],
                },
            }
        )

        self.assertEqual("src/app/old.py", request.old_path)
        self.assertTrue(request.dry_run)
        self.assertTrue(request.moves[0].tracked)
        self.assertTrue(request.options.include_python)
        self.assertEqual(("src/app/generated/**",), request.options.scan_exclude)

    def test_response_emits_core_protocol_shape(self) -> None:
        response = AdapterResponse(
            symbol_mappings=(
                SymbolMapping(
                    kind="module",
                    old_path="src/app/old.py",
                    new_path="src/app/new.py",
                    old_symbol="app.old",
                    new_symbol="app.new",
                ),
            ),
            replacements=(
                Replacement(
                    file="src/app/use.py",
                    start=1,
                    end=8,
                    replacement="app.new",
                    reason="python-import",
                    rule="ImportStatementReplacementRule",
                ),
            ),
            warnings=(Warning(message="dynamic import detected", file="src/app/use.py", line=4),),
        )

        payload = response.to_payload()

        self.assertEqual(1, payload["protocolVersion"])
        self.assertEqual("python", payload["adapter"])
        self.assertEqual("app.old", payload["symbolMappings"][0]["oldSymbol"])
        self.assertEqual("python-import", payload["replacements"][0]["reason"])
        self.assertEqual(4, payload["warnings"][0]["line"])

    def test_analyze_command_rejects_invalid_protocol(self) -> None:
        stdout = io.StringIO()
        stderr = io.StringIO()

        code = run(["refactorlah-python", "analyze"], io.StringIO("{}"), stdout, stderr)

        self.assertEqual(1, code)
        self.assertIn("protocolVersion", stderr.getvalue())
        self.assertIn("errors", json.loads(stdout.getvalue()))


if __name__ == "__main__":
    unittest.main()
