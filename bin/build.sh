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
DIST_DIR=$BUILD_DIR/dist
HOST_BINARY=$BUILD_DIR/refactorlah
BUILD_README=$BUILD_DIR/README.txt
GO_CACHE_DIR=$ROOT_DIR/.cache/go-build
DEFAULT_TARGETS="darwin/arm64 linux/arm64 windows/arm64"
RUN_TESTS=1
TARGET_MODE=host
TARGETS=

usage() {
  cat <<'EOF'
usage: bin/build.sh [options]

Build refactorlah CLI bundles.

Options:
  --target host          Build the current GOOS/GOARCH target (default)
  --target all           Build the release target matrix
  --target GOOS/GOARCH   Build one explicit target, for example linux/arm64
  --all                  Alias for --target all
  --no-test              Skip the pre-build Go test suite
  -h, --help             Show this help

Notes:
  Built-in PHP/Python analysers use cgo through tree-sitter. Cross-target builds
  need a working C compiler for that target, or should be run on that target OS.
EOF
}

add_target() {
  target=$1
  refactorlah_validate_target "$target"
  case " $TARGETS " in
    *" $target "*) ;;
    *) TARGETS="${TARGETS}${TARGETS:+ }$target" ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --target)
      shift
      if [ "$#" -eq 0 ]; then
        echo "error: --target requires a value" >&2
        usage >&2
        exit 2
      fi
      case "$1" in
        host) TARGET_MODE=host ;;
        all) TARGET_MODE=all ;;
        *) TARGET_MODE=custom; add_target "$1" ;;
      esac
      ;;
    --target=*)
      value=${1#--target=}
      case "$value" in
        host) TARGET_MODE=host ;;
        all) TARGET_MODE=all ;;
        *) TARGET_MODE=custom; add_target "$value" ;;
      esac
      ;;
    --all)
      TARGET_MODE=all
      ;;
    --no-test)
      RUN_TESTS=0
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
  shift
done

refactorlah_require_command go "Building the Go CLI"

HOST_TARGET=$(refactorlah_host_target)
case "$TARGET_MODE" in
  host) TARGETS=$HOST_TARGET ;;
  all) TARGETS=$DEFAULT_TARGETS ;;
esac

if [ "$RUN_TESTS" -eq 1 ]; then
  echo "Running test suite before build"
  "$ROOT_DIR/bin/test.sh"
fi

echo "Building refactorlah into $BUILD_DIR"
refactorlah_remove_directory "$BUILD_DIR"
mkdir -p "$DIST_DIR"
mkdir -p "$GO_CACHE_DIR"

write_bundle_readme() {
  readme_path=$1
  binary_name=$2
  target=$3

  cat > "$readme_path" <<EOF
refactorlah build bundle
========================

Target:
- $target

Contents:
- $binary_name

Normal usage:
  cd /path/to/target-project
  ./$binary_name move old/path new/path

Examples:
  ./$binary_name move app/Services/Billing app/Domain/Billing
  ./$binary_name move src/app/services/billing.py src/app/domain/billing.py
  ./$binary_name move app/Services/Billing app/Domain/Billing --dry
  ./$binary_name move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php

Notes:
- Apply is the default. Use --dry to preview changes.
- PHP, Python, Go, Symfony/Twig, and static import analysis are built into the CLI.
- PHP and Python runtimes are not required when using this built binary.
- This bundle is source-checkout-independent and does not depend on the repository after install.
EOF
}

build_target() {
  target=$1
  goos=$(refactorlah_target_goos "$target")
  goarch=$(refactorlah_target_goarch "$target")
  slug=$(refactorlah_target_slug "$target")
  bundle_dir=$DIST_DIR/refactorlah_$slug
  binary_name=$(refactorlah_binary_name "$target")
  binary_path=$bundle_dir/$binary_name

  mkdir -p "$bundle_dir"

  echo "Building CLI for $target"
  if ! (
    cd "$ROOT_DIR"
    env CGO_ENABLED=1 GOOS="$goos" GOARCH="$goarch" GOCACHE="${GOCACHE:-$GO_CACHE_DIR}" go build -o "$binary_path" ./cmd/refactorlah
  ); then
    cat >&2 <<EOF
error: failed to build target $target

Built-in PHP/Python support is compiled through cgo. Cross-target builds require
a C compiler/toolchain for $target. If you do not have one locally, build this
target on a matching OS/architecture runner instead.
EOF
    exit 1
  fi

  chmod +x "$binary_path" 2>/dev/null || true
  write_bundle_readme "$bundle_dir/README.txt" "$binary_name" "$target"

  if [ "$target" = "$HOST_TARGET" ]; then
    cp "$binary_path" "$HOST_BINARY"
    chmod +x "$HOST_BINARY" 2>/dev/null || true
    cp "$bundle_dir/README.txt" "$BUILD_README"
  fi
}

for target in $TARGETS; do
  build_target "$target"
done

cat > "$BUILD_DIR/targets.txt" <<EOF
Targets built:
$(printf '%s\n' $TARGETS)
EOF

cat <<EOF

Build complete.

Built targets:
$(printf '  %s\n' $TARGETS)

Bundles:
  $DIST_DIR
EOF

if [ -x "$HOST_BINARY" ]; then
  cat <<EOF

Host command:
  $HOST_BINARY

Example:
  cd /path/to/target-project
  $HOST_BINARY move old/path new/path
EOF
fi
