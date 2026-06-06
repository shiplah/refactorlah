#!/usr/bin/env bash

if [[ -z "${BASH_VERSION:-}" ]]; then
  echo "error: bin/install.sh must be run with bash. Try: bash bin/install.sh" >&2
  exit 2
fi

set -euo pipefail

SCRIPT_SOURCE="${BASH_SOURCE[0]:-$0}"
ROOT_DIR="$(cd "$(dirname "${SCRIPT_SOURCE}")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
BUILD_BINARY="${ROOT_DIR}/build/refactorlah"

if [[ -n "${2:-}" ]]; then
  echo "error: too many arguments" >&2
  echo "usage: bin/install.sh [install-dir]" >&2
  exit 2
fi

INSTALL_DIR="${1:-}"
if [[ -z "${INSTALL_DIR}" ]]; then
  if [[ -z "${HOME:-}" ]]; then
    echo "error: install directory not provided and HOME is not set" >&2
    echo "usage: bin/install.sh /path/to/bin" >&2
    exit 2
  fi
  INSTALL_DIR="${HOME}/.local/bin"
fi

mkdir -p "${INSTALL_DIR}"
INSTALL_DIR="$(cd "${INSTALL_DIR}" && pwd)"
BUNDLE_DIR="${INSTALL_DIR}/refactorlah.bundle"
TARGET_LINK="${INSTALL_DIR}/refactorlah"
case "${BUNDLE_DIR}/" in
  "${BUILD_DIR}/"*)
    echo "error: install directory must not be inside ${BUILD_DIR}" >&2
    exit 2
    ;;
esac

remove_directory() {
  local path="$1"

  if [[ -z "${path}" || "${path}" == "/" ]]; then
    echo "error: refusing to remove unsafe path '${path}'" >&2
    exit 2
  fi

  rm -rf "${path}"
}

echo "Building refactorlah bundle"
"${ROOT_DIR}/bin/build.sh"

if [[ -e "${TARGET_LINK}" && ! -L "${TARGET_LINK}" ]]; then
  echo "error: ${TARGET_LINK} already exists and is not a symlink" >&2
  exit 2
fi

if [[ ! -x "${BUILD_BINARY}" ]]; then
  echo "error: build binary missing at ${BUILD_BINARY}" >&2
  exit 2
fi

remove_directory "${BUNDLE_DIR}"
mkdir -p "${BUNDLE_DIR}"
cp -R "${BUILD_DIR}/." "${BUNDLE_DIR}/"
ln -sfn "${BUNDLE_DIR}/refactorlah" "${TARGET_LINK}"

cat <<EOF
Installed refactorlah symlink:
  ${TARGET_LINK} -> ${BUNDLE_DIR}/refactorlah

Installed bundle:
  ${BUNDLE_DIR}

If ${INSTALL_DIR} is not already on your PATH, add it in your shell profile.
EOF
