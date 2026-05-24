#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
GO_BINARY="${BUILD_DIR}/refactorlah"
PHP_ADAPTER_SOURCE_DIR="${ROOT_DIR}/adapters/php"
PHP_ADAPTER_BUILD_DIR="${BUILD_DIR}/libexec/refactorlah-php"
PYTHON_ADAPTER_SOURCE_DIR="${ROOT_DIR}/adapters/python"
PYTHON_ADAPTER_BUILD_DIR="${BUILD_DIR}/libexec/refactorlah-python"
BUILD_README="${BUILD_DIR}/README.txt"
GO_CACHE_DIR="${ROOT_DIR}/.cache/go-build"

echo "Building refactorlah into ${BUILD_DIR}"

echo "Running test suite before build"
"${ROOT_DIR}/bin/test.sh"

rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"
mkdir -p "${PHP_ADAPTER_BUILD_DIR}"
mkdir -p "${PYTHON_ADAPTER_BUILD_DIR}"
mkdir -p "${GO_CACHE_DIR}"

echo "Building Go CLI"
(
  cd "${ROOT_DIR}"
  GOCACHE="${GOCACHE:-${GO_CACHE_DIR}}" go build -o "${GO_BINARY}" ./cmd/refactorlah
)

echo "Staging PHP adapter"
cp -R "${PHP_ADAPTER_SOURCE_DIR}/bin" "${PHP_ADAPTER_BUILD_DIR}/"
cp -R "${PHP_ADAPTER_SOURCE_DIR}/src" "${PHP_ADAPTER_BUILD_DIR}/"
cp "${PHP_ADAPTER_SOURCE_DIR}/composer.json" "${PHP_ADAPTER_BUILD_DIR}/composer.json"
cp "${PHP_ADAPTER_SOURCE_DIR}/composer.lock" "${PHP_ADAPTER_BUILD_DIR}/composer.lock"

if [[ -d "${PHP_ADAPTER_SOURCE_DIR}/vendor" ]]; then
  echo "Copying existing Composer dependencies"
  cp -R "${PHP_ADAPTER_SOURCE_DIR}/vendor" "${PHP_ADAPTER_BUILD_DIR}/"
else
  echo "Installing Composer dependencies into build bundle"
  (
    cd "${PHP_ADAPTER_BUILD_DIR}"
    composer install --no-dev --prefer-dist --no-progress
  )
fi

echo "Staging Python adapter"
cp -R "${PYTHON_ADAPTER_SOURCE_DIR}/bin" "${PYTHON_ADAPTER_BUILD_DIR}/"
cp -R "${PYTHON_ADAPTER_SOURCE_DIR}/src" "${PYTHON_ADAPTER_BUILD_DIR}/"
cp "${PYTHON_ADAPTER_SOURCE_DIR}/pyproject.toml" "${PYTHON_ADAPTER_BUILD_DIR}/pyproject.toml"

chmod +x "${GO_BINARY}"
chmod +x "${PHP_ADAPTER_BUILD_DIR}/bin/refactorlah-php"
chmod +x "${PYTHON_ADAPTER_BUILD_DIR}/bin/refactorlah-python"

cat > "${BUILD_README}" <<EOF
refactorlah build bundle
========================

Contents:
- refactorlah
- libexec/refactorlah-php/
- libexec/refactorlah-python/

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
- The CLI auto-discovers bundled adapters in ./libexec/.
- PHP must be available on the machine when PHP refactors are executed.
- Python 3 must be available on the machine when Python refactors are executed.
- This bundle is self-contained and does not depend on the source checkout at runtime.
EOF

cat <<EOF

Build complete.

User-facing command:
  ${GO_BINARY}

Bundled PHP adapter:
  ${PHP_ADAPTER_BUILD_DIR}/bin/refactorlah-php

Bundled Python adapter:
  ${PYTHON_ADAPTER_BUILD_DIR}/bin/refactorlah-python

Bundle README:
  ${BUILD_README}

Example:
  cd /path/to/target-project
  ${GO_BINARY} move old/path new/path
EOF
