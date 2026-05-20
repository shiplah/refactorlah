#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_BINARY="${ROOT_DIR}/build/refactorlah"
INSTALL_DIR="${1:-${HOME}/.local/bin}"
TARGET_LINK="${INSTALL_DIR}/refactorlah"

echo "Building refactorlah bundle"
"${ROOT_DIR}/bin/build.sh"

mkdir -p "${INSTALL_DIR}"
ln -sfn "${BUILD_BINARY}" "${TARGET_LINK}"

cat <<EOF
Installed refactorlah symlink:
  ${TARGET_LINK} -> ${BUILD_BINARY}

If ${INSTALL_DIR} is not already on your PATH, add it in your shell profile.
EOF
