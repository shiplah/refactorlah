#!/bin/sh

refactorlah_absolute_dir() {
  if [ -z "$1" ]; then
    echo "error: missing directory" >&2
    return 2
  fi

  CDPATH= cd -- "$1" && pwd
}

refactorlah_require_command() {
  command_name=$1
  purpose=$2

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "error: $purpose requires '$command_name' on PATH" >&2
    exit 127
  fi
}

refactorlah_remove_directory() {
  remove_path=$1

  if [ -z "$remove_path" ] || [ "$remove_path" = "/" ] || [ "$remove_path" = "." ] || [ "$remove_path" = ".." ]; then
    echo "error: refusing to remove unsafe path '$remove_path'" >&2
    exit 2
  fi

  rm -rf "$remove_path"
}

refactorlah_host_target() {
  goos=$(go env GOOS)
  goarch=$(go env GOARCH)
  printf '%s/%s\n' "$goos" "$goarch"
}

refactorlah_target_goos() {
  printf '%s\n' "$1" | awk -F/ '{print $1}'
}

refactorlah_target_goarch() {
  printf '%s\n' "$1" | awk -F/ '{print $2}'
}

refactorlah_validate_target() {
  target=$1
  goos=$(refactorlah_target_goos "$target")
  goarch=$(refactorlah_target_goarch "$target")
  extra=$(printf '%s\n' "$target" | awk -F/ '{print $3}')

  if [ -z "$goos" ] || [ -z "$goarch" ] || [ -n "$extra" ]; then
    echo "error: invalid target '$target'; expected GOOS/GOARCH, for example darwin/arm64" >&2
    exit 2
  fi
}

refactorlah_target_slug() {
  printf '%s\n' "$1" | tr '/' '-'
}

refactorlah_binary_name() {
  goos=$(refactorlah_target_goos "$1")
  if [ "$goos" = "windows" ]; then
    printf 'refactorlah.exe\n'
    return
  fi

  printf 'refactorlah\n'
}

refactorlah_path_contains() {
  needle=$1
  path_value=${2:-}

  if [ -z "$needle" ] || [ -z "$path_value" ]; then
    return 1
  fi

  old_ifs=$IFS
  IFS=:
  found=1
  for entry in $path_value; do
    if [ "$entry" = "$needle" ]; then
      found=0
      break
    fi
  done
  IFS=$old_ifs

  return "$found"
}
