#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_BINARY="${ROOT_DIR}/build/refactorlah"
INSTALL_DIR="${1:-${HOME}/.local/bin}"
TARGET_LINK="${INSTALL_DIR}/refactorlah"

if [[ ! -x "${BUILD_BINARY}" ]]; then
  echo "Built binary not found at ${BUILD_BINARY}" >&2
  echo "Run bin/build.sh first." >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}"
ln -sfn "${BUILD_BINARY}" "${TARGET_LINK}"

cat <<EOF
Installed refactorlah symlink:
  ${TARGET_LINK} -> ${BUILD_BINARY}

If ${INSTALL_DIR} is not already on your PATH, add it in your shell profile.
EOF
