#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_CACHE_DIR="${ROOT_DIR}/.cache/go-build"
PHP_ADAPTER_DIR="${ROOT_DIR}/adapters/php"
PYTHON_ADAPTER_DIR="${ROOT_DIR}/adapters/python"

mkdir -p "${GO_CACHE_DIR}"

echo "Running Go tests"
(
  cd "${ROOT_DIR}"
  GOCACHE="${GOCACHE:-${GO_CACHE_DIR}}" go test ./...
)

echo
echo "Running PHP adapter tests"
(
  cd "${PHP_ADAPTER_DIR}"
  composer test
)

echo
echo "Running Python adapter tests"
(
  cd "${PYTHON_ADAPTER_DIR}"
  python3 tests/run.py
)

echo
echo "All tests passed."
