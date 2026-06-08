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
refactorlah_require_command gofmt "Go formatting"

echo "Running shell script tests"
for script in "$ROOT_DIR"/bin/*.sh; do
  sh -n "$script"
done

if ! refactorlah_path_contains "/example/bin" "/usr/bin:/example/bin:/bin"; then
  echo "error: expected PATH helper to find existing entry" >&2
  exit 1
fi
if refactorlah_path_contains "/example/bin" "/usr/bin:/example/binary:/bin"; then
  echo "error: PATH helper matched partial entry" >&2
  exit 1
fi
original_ifs=$IFS
if ! refactorlah_path_contains "/example/bin" "/example/bin"; then
  echo "error: expected PATH helper to find single entry" >&2
  exit 1
fi
if [ "$IFS" != "$original_ifs" ]; then
  echo "error: PATH helper did not restore IFS" >&2
  exit 1
fi

"$ROOT_DIR/bin/build.sh" --help >/dev/null
"$ROOT_DIR/bin/format.sh" --help >/dev/null
"$ROOT_DIR/bin/install.sh" --help >/dev/null
"$ROOT_DIR/bin/test.sh" --help >/dev/null

echo "Checking Go formatting"
unformatted=$(find "$ROOT_DIR" \
  -name '*.go' \
  -not -path "$ROOT_DIR/.cache/*" \
  -not -path "$ROOT_DIR/build/*" \
  -not -path "$ROOT_DIR/tests/fixtures/*" \
  -exec gofmt -l {} +)
if [ -n "$unformatted" ]; then
  echo "error: Go files need gofmt:" >&2
  printf '%s\n' "$unformatted" >&2
  exit 1
fi

echo "Running Go tests"
mkdir -p "$GO_CACHE_DIR"
(
  cd "$ROOT_DIR"
  GOCACHE="${GOCACHE:-$GO_CACHE_DIR}" go test ./...
)

echo
echo "All tests passed."
