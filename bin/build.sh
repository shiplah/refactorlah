#!/usr/bin/env bash

if [[ -z "${BASH_VERSION:-}" ]]; then
  echo "error: bin/build.sh must be run with bash. Try: bash bin/build.sh" >&2
  exit 2
fi

set -euo pipefail

SCRIPT_SOURCE="${BASH_SOURCE[0]:-$0}"
ROOT_DIR="$(cd "$(dirname "${SCRIPT_SOURCE}")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
GO_BINARY="${BUILD_DIR}/refactorlah"
BUILD_README="${BUILD_DIR}/README.txt"
GO_CACHE_DIR="${ROOT_DIR}/.cache/go-build"

require_command() {
  local command_name="$1"
  local purpose="$2"

  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "error: ${purpose} requires '${command_name}' on PATH" >&2
    exit 127
  fi
}

remove_directory() {
  local path="$1"

  if [[ -z "${path}" || "${path}" == "/" ]]; then
    echo "error: refusing to remove unsafe path '${path}'" >&2
    exit 2
  fi

  rm -rf "${path}"
}

require_command go "Building the Go CLI"

echo "Building refactorlah into ${BUILD_DIR}"

echo "Running test suite before build"
"${ROOT_DIR}/bin/test.sh" --go-only

remove_directory "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"
mkdir -p "${GO_CACHE_DIR}"

echo "Building Go CLI with native analyzers"
(
  cd "${ROOT_DIR}"
  CGO_ENABLED=1 GOCACHE="${GOCACHE:-${GO_CACHE_DIR}}" go build -o "${GO_BINARY}" ./cmd/refactorlah
)

chmod +x "${GO_BINARY}"

cat > "${BUILD_README}" <<EOF
refactorlah build bundle
========================

Contents:
- refactorlah

Normal usage:
  cd /path/to/target-project
  $(basename "${GO_BINARY}") move old/path new/path

Examples:
  ./refactorlah move app/Services/Billing app/Domain/Billing
  ./refactorlah move src/app/services/billing.py src/app/domain/billing.py
  ./refactorlah move app/Services/Billing app/Domain/Billing --dry
  ./refactorlah move --use-list app/Foo.php,app/Bar.php tests/A.php,tests/B.php

Notes:
- Apply is the default. Use --dry to preview changes.
- PHP, Python, Go, Symfony/Twig, and static import analysis are built into the CLI.
- PHP and Python runtimes are not required when using this built binary.
- This bundle is source-checkout-independent and does not depend on the repository after install.
EOF

cat <<EOF

Build complete.

User-facing command:
  ${GO_BINARY}

Bundle README:
  ${BUILD_README}

Example:
  cd /path/to/target-project
  ${GO_BINARY} move old/path new/path
EOF
