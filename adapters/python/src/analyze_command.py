from __future__ import annotations

import json
import sys
from typing import Sequence, TextIO

from src.protocol.request import AdapterRequest
from src.protocol.response import AdapterResponse


def run(argv: Sequence[str], stdin: TextIO, stdout: TextIO, stderr: TextIO) -> int:
    if len(argv) < 2 or argv[1] != "analyze":
        stderr.write("usage: refactorlah-python analyze\n")
        return 2

    try:
        payload = json.loads(stdin.read())
        if not isinstance(payload, dict):
            raise ValueError("adapter request must decode to an object")

        AdapterRequest.from_payload(payload)
        response = AdapterResponse()
        stdout.write(json.dumps(response.to_payload(), indent=2))
        return 0
    except Exception as exc:
        stderr.write(f"{exc}\n")
        response = AdapterResponse(errors=(str(exc),))
        stdout.write(json.dumps(response.to_payload(), indent=2))
        return 1


def main(argv: Sequence[str] | None = None) -> int:
    return run(sys.argv if argv is None else argv, sys.stdin, sys.stdout, sys.stderr)
