#!/bin/sh

set -eu

case "$0" in
  */*) script_dir=$(dirname "$0") ;;
  *) script_dir=$(dirname "$(command -v "$0")") ;;
esac
SCRIPT_DIR=$(CDPATH= cd -- "$script_dir" && pwd)
. "$SCRIPT_DIR/_lib.sh"
ROOT_DIR=$(refactorlah_absolute_dir "$SCRIPT_DIR/..")

usage() {
  cat <<'EOF'
usage: bin/format.sh

Format Go source files with gofmt.
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

refactorlah_require_command gofmt "Go formatting"

find "$ROOT_DIR" \
  -name '*.go' \
  -not -path "$ROOT_DIR/.cache/*" \
  -not -path "$ROOT_DIR/build/*" \
  -not -path "$ROOT_DIR/tests/fixtures/*" \
  -exec gofmt -w {} +
