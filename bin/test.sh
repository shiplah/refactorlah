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

usage() {
  cat <<'EOF'
usage: bin/test.sh
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

refactorlah_require_command go "Go tests"

mkdir -p "$GO_CACHE_DIR"

echo "Running Go tests"
(
  cd "$ROOT_DIR"
  GOCACHE="${GOCACHE:-$GO_CACHE_DIR}" go test ./...
)

echo
echo "All tests passed."
