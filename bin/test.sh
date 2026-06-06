#!/bin/sh

set -eu

case "$0" in
  */*) script_dir=$(dirname "$0") ;;
  *) script_dir=$(dirname "$(command -v "$0")") ;;
esac
SCRIPT_DIR=$(CDPATH= cd -- "$script_dir" && pwd)
. "$SCRIPT_DIR/_lib.sh"
ROOT_DIR=$(refactorlah_absolute_dir "$SCRIPT_DIR/..")

GO_CACHE_DIR=$ROOT_DIR/.cache/go-build
PHP_ADAPTER_DIR=$ROOT_DIR/adapters/php
PYTHON_ADAPTER_DIR=$ROOT_DIR/adapters/python
GO_ONLY=0

usage() {
  cat <<'EOF'
usage: bin/test.sh [--go-only]
EOF
}

if [ "$#" -gt 1 ]; then
  echo "error: too many arguments" >&2
  usage >&2
  exit 2
fi

case "${1:-}" in
  "")
    ;;
  --go-only)
    GO_ONLY=1
    ;;
  -h|--help)
    usage
    exit 0
    ;;
  *)
    echo "error: unknown option '$1'" >&2
    usage >&2
    exit 2
    ;;
esac

refactorlah_require_command go "Go tests"
if [ "$GO_ONLY" -eq 0 ]; then
  refactorlah_require_command php "PHP adapter tests"
  refactorlah_require_command composer "PHP adapter tests"
  refactorlah_require_command python3 "Python adapter tests"

  if ! php -r "exit(version_compare(PHP_VERSION, '8.2.0', '>=') ? 0 : 1);" >/dev/null 2>&1; then
    echo "error: PHP adapter tests require php >= 8.2" >&2
    exit 127
  fi

  if ! python3 -c 'import sys; raise SystemExit(0 if sys.version_info >= (3, 11) else 1)' >/dev/null 2>&1; then
    echo "error: Python adapter tests require python3 >= 3.11" >&2
    echo "hint: ensure a compatible python3 appears before older system Python on PATH" >&2
    exit 127
  fi
fi

mkdir -p "$GO_CACHE_DIR"

echo "Running Go tests"
(
  cd "$ROOT_DIR"
  GOCACHE="${GOCACHE:-$GO_CACHE_DIR}" go test ./...
)

if [ "$GO_ONLY" -eq 1 ]; then
  echo
  echo "Go tests passed."
  exit 0
fi

echo
echo "Running PHP adapter tests"
(
  cd "$PHP_ADAPTER_DIR"
  if [ ! -d vendor ]; then
    echo "Installing PHP adapter Composer dependencies"
    composer install --no-progress
  fi
  composer test
)

echo
echo "Running Python adapter tests"
(
  cd "$PYTHON_ADAPTER_DIR"
  python3 tests/run.py
)

echo
echo "All tests passed."
