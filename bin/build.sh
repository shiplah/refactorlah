#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
GO_BINARY="${BUILD_DIR}/refactorlah"
ADAPTER_SOURCE_DIR="${ROOT_DIR}/adapters/php"
ADAPTER_BUILD_DIR="${BUILD_DIR}/libexec/refactorlah-php"
BUILD_README="${BUILD_DIR}/README.txt"
GO_CACHE_DIR="${ROOT_DIR}/.cache/go-build"

echo "Building refactorlah into ${BUILD_DIR}"

rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"
mkdir -p "${ADAPTER_BUILD_DIR}"
mkdir -p "${GO_CACHE_DIR}"

echo "Building Go CLI"
(
  cd "${ROOT_DIR}"
  GOCACHE="${GOCACHE:-${GO_CACHE_DIR}}" go build -o "${GO_BINARY}" ./cmd/refactorlah
)

echo "Staging PHP adapter"
cp -R "${ADAPTER_SOURCE_DIR}/bin" "${ADAPTER_BUILD_DIR}/"
cp -R "${ADAPTER_SOURCE_DIR}/src" "${ADAPTER_BUILD_DIR}/"
cp "${ADAPTER_SOURCE_DIR}/composer.json" "${ADAPTER_BUILD_DIR}/composer.json"
cp "${ADAPTER_SOURCE_DIR}/composer.lock" "${ADAPTER_BUILD_DIR}/composer.lock"

if [[ -d "${ADAPTER_SOURCE_DIR}/vendor" ]]; then
  echo "Copying existing Composer dependencies"
  cp -R "${ADAPTER_SOURCE_DIR}/vendor" "${ADAPTER_BUILD_DIR}/"
else
  echo "Installing Composer dependencies into build bundle"
  (
    cd "${ADAPTER_BUILD_DIR}"
    composer install --no-dev --prefer-dist --no-progress
  )
fi

chmod +x "${GO_BINARY}"
chmod +x "${ADAPTER_BUILD_DIR}/bin/refactorlah-php"

cat > "${BUILD_README}" <<EOF
refactorlah build bundle
========================

Contents:
- refactorlah
- libexec/refactorlah-php/

Normal usage:
  cd /path/to/target-project
  $(basename "${GO_BINARY}") old/path new/path

Examples:
  ./refactorlah app/Services/Billing app/Domain/Billing
  ./refactorlah app/Services/Billing app/Domain/Billing --apply

Notes:
- Dry-run is the default.
- The CLI auto-discovers the bundled PHP adapter in ./libexec/refactorlah-php/.
- PHP must be available on the machine when PHP refactors are executed.
- This bundle is self-contained and does not depend on the source checkout at runtime.
EOF

cat <<EOF

Build complete.

User-facing command:
  ${GO_BINARY}

Bundled PHP adapter:
  ${ADAPTER_BUILD_DIR}/bin/refactorlah-php

Bundle README:
  ${BUILD_README}

Example:
  cd ~/Code/example/project
  ${GO_BINARY} old/path new/path
EOF
