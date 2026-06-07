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
BUILD_README=$BUILD_DIR/README.txt

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
BUNDLE_DIR=$INSTALL_DIR/refactorlah.bundle
TARGET_LINK=$INSTALL_DIR/refactorlah

case "$BUNDLE_DIR/" in
  "$BUILD_DIR/"*)
    echo "error: install directory must not be inside $BUILD_DIR" >&2
    exit 2
    ;;
esac

echo "Installing refactorlah into $INSTALL_DIR"
echo
"$ROOT_DIR/bin/build.sh" --target host --no-summary

if [ -e "$TARGET_LINK" ] && [ ! -L "$TARGET_LINK" ]; then
  echo "error: $TARGET_LINK already exists and is not a symlink" >&2
  exit 2
fi

if [ ! -x "$BUILD_BINARY" ]; then
  echo "error: build binary missing at $BUILD_BINARY" >&2
  exit 2
fi

refactorlah_remove_directory "$BUNDLE_DIR"
mkdir -p "$BUNDLE_DIR"
cp "$BUILD_BINARY" "$BUNDLE_DIR/refactorlah"
if [ -f "$BUILD_README" ]; then
  cp "$BUILD_README" "$BUNDLE_DIR/README.txt"
fi
rm -f "$TARGET_LINK"
ln -s "$BUNDLE_DIR/refactorlah" "$TARGET_LINK"

cat <<EOF

Install complete.

Command:
  $TARGET_LINK

Bundle:
  $BUNDLE_DIR
EOF

if ! refactorlah_path_contains "$INSTALL_DIR" "${PATH:-}"; then
  cat <<EOF
If $INSTALL_DIR is not already on your PATH, add it in your shell profile.
EOF
fi
