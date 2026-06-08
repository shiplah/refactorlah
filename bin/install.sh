#!/bin/sh

set -eu

case "$0" in
  */*) script_dir=$(dirname "$0") ;;
  *) script_dir=$(dirname "$(command -v "$0")") ;;
esac
SCRIPT_DIR=$(CDPATH= cd -- "$script_dir" && pwd)
. "$SCRIPT_DIR/_lib.sh"
ROOT_DIR=$(refactorlah_absolute_dir "$SCRIPT_DIR/..")

BUILD_DIR=$ROOT_DIR/build
BUILD_BINARY=$BUILD_DIR/refactorlah
tmp_binary=

cleanup_tmp_binary() {
  if [ -n "${tmp_binary:-}" ] && [ -f "$tmp_binary" ]; then
    rm -f "$tmp_binary"
  fi
}

trap cleanup_tmp_binary EXIT HUP INT TERM

usage() {
  cat <<'EOF'
usage: bin/install.sh [install-dir]

Build and install the host refactorlah binary.
EOF
}

if [ "$#" -gt 1 ]; then
  echo "error: too many arguments" >&2
  usage >&2
  exit 2
fi

case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
esac

INSTALL_DIR=${1:-}
if [ -z "$INSTALL_DIR" ]; then
  if [ -z "${HOME:-}" ]; then
    echo "error: install directory not provided and HOME is not set" >&2
    usage >&2
    exit 2
  fi
  INSTALL_DIR=$HOME/.local/bin
fi

mkdir -p "$INSTALL_DIR"
INSTALL_DIR=$(refactorlah_absolute_dir "$INSTALL_DIR")
TARGET_BINARY=$INSTALL_DIR/refactorlah
STALE_BUNDLE_DIR=$INSTALL_DIR/refactorlah.bundle

case "$INSTALL_DIR/" in
  "$BUILD_DIR/"*)
    echo "error: install directory must not be inside $BUILD_DIR" >&2
    exit 2
    ;;
esac

echo "Installing refactorlah into $INSTALL_DIR"
echo
REFACTORLAH_DISTRIBUTION=source-install "$ROOT_DIR/bin/build.sh" --target host --no-summary

if [ ! -x "$BUILD_BINARY" ]; then
  echo "error: build binary missing at $BUILD_BINARY" >&2
  exit 2
fi

if [ -d "$TARGET_BINARY" ]; then
  echo "error: $TARGET_BINARY already exists and is a directory" >&2
  exit 2
fi

tmp_binary=$INSTALL_DIR/.refactorlah.tmp.$$
rm -f "$tmp_binary"
cp "$BUILD_BINARY" "$tmp_binary"
chmod +x "$tmp_binary" 2>/dev/null || true
mv -f "$tmp_binary" "$TARGET_BINARY"
tmp_binary=

if [ -d "$STALE_BUNDLE_DIR" ]; then
  refactorlah_remove_directory "$STALE_BUNDLE_DIR"
fi

cat <<EOF

Install complete.

Command:
  $TARGET_BINARY
EOF

if ! refactorlah_path_contains "$INSTALL_DIR" "${PATH:-}"; then
  cat <<EOF
If $INSTALL_DIR is not already on your PATH, add it in your shell profile.
EOF
fi
