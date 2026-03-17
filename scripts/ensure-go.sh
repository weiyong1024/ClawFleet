#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH='' cd -- "$SCRIPT_DIR/.." && pwd)
GO_MOD_FILE="$REPO_ROOT/go.mod"

usage() {
  cat >&2 <<'EOF'
Usage: ensure-go.sh --print-path

Ensures the Go toolchain required by go.mod is available. If Go is missing or
too old, downloads the official archive from go.dev into a user-local toolchain
directory and prints the resolved go binary path.
EOF
}

log() {
  printf '==> %s\n' "$*" >&2
}

die() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

required_go_version() {
  awk '/^go / { print $2; exit }' "$GO_MOD_FILE"
}

version_parts() {
  version=${1#go}
  old_ifs=$IFS
  IFS=.
  set -- $version
  IFS=$old_ifs
  printf '%s %s %s\n' "${1:-0}" "${2:-0}" "${3:-0}"
}

version_ge() {
  left_version=$1
  right_version=$2

  set -- $(version_parts "$left_version")
  left_major=$1
  left_minor=$2
  left_patch=$3

  set -- $(version_parts "$right_version")
  right_major=$1
  right_minor=$2
  right_patch=$3

  if [ "$left_major" -gt "$right_major" ]; then
    return 0
  fi
  if [ "$left_major" -lt "$right_major" ]; then
    return 1
  fi

  if [ "$left_minor" -gt "$right_minor" ]; then
    return 0
  fi
  if [ "$left_minor" -lt "$right_minor" ]; then
    return 1
  fi

  [ "$left_patch" -ge "$right_patch" ]
}

toolchain_root() {
  if [ -n "${CLAWFLEET_TOOLCHAIN_HOME:-}" ]; then
    printf '%s\n' "$CLAWFLEET_TOOLCHAIN_HOME"
    return
  fi

  case "$(uname -s)" in
    Darwin)
      printf '%s\n' "$HOME/Library/Application Support/ClawFleet/toolchains"
      ;;
    Linux)
      printf '%s\n' "${XDG_DATA_HOME:-$HOME/.local/share}/clawfleet/toolchains"
      ;;
    *)
      die "unsupported OS: $(uname -s)"
      ;;
  esac
}

go_os() {
  case "$(uname -s)" in
    Darwin) printf 'darwin\n' ;;
    Linux) printf 'linux\n' ;;
    *) die "unsupported OS: $(uname -s)" ;;
  esac
}

go_arch() {
  case "$(uname -m)" in
    arm64|aarch64) printf 'arm64\n' ;;
    x86_64|amd64) printf 'amd64\n' ;;
    *) die "unsupported architecture: $(uname -m)" ;;
  esac
}

download_file() {
  url=$1
  destination=$2

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$destination"
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$destination" "$url"
    return
  fi

  die "curl or wget is required to download Go from $url"
}

go_version_at() {
  go_bin=$1
  "$go_bin" version | awk '{ print $3 }'
}

resolve_existing_go() {
  go_bin=$1
  if [ ! -x "$go_bin" ]; then
    return 1
  fi

  current_version=$(go_version_at "$go_bin")
  if version_ge "$current_version" "$REQUIRED_GO_VERSION"; then
    printf '%s\n' "$go_bin"
    return 0
  fi

  return 1
}

install_go() {
  install_root=$1
  archive_name="go${REQUIRED_GO_VERSION}.$(go_os)-$(go_arch).tar.gz"
  download_url="https://go.dev/dl/$archive_name"

  tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/clawfleet-go.XXXXXX")
  trap 'rm -rf "$tmp_dir"' EXIT INT TERM HUP

  archive_path="$tmp_dir/$archive_name"
  staging_dir="$tmp_dir/staging"
  target_tmp="$install_root.tmp.$$"

  log "Go ${REQUIRED_GO_VERSION} is required but was not found. Downloading $download_url ..."
  download_file "$download_url" "$archive_path"

  mkdir -p "$staging_dir"
  tar -C "$staging_dir" -xzf "$archive_path"
  [ -x "$staging_dir/go/bin/go" ] || die "downloaded archive did not contain a Go toolchain"

  mkdir -p "$(dirname "$install_root")"
  rm -rf "$target_tmp"
  mv "$staging_dir/go" "$target_tmp"
  rm -rf "$install_root"
  mv "$target_tmp" "$install_root"

  trap - EXIT INT TERM HUP
  rm -rf "$tmp_dir"

  log "Installed Go ${REQUIRED_GO_VERSION} to $install_root"
  log "Add \"$install_root/bin\" to PATH if you want to use this toolchain outside Make."
}

main() {
  if [ "${1:-}" != "--print-path" ] || [ "$#" -ne 1 ]; then
    usage
    exit 2
  fi

  [ -f "$GO_MOD_FILE" ] || die "go.mod not found at $GO_MOD_FILE"
  REQUIRED_GO_VERSION=$(required_go_version)
  [ -n "$REQUIRED_GO_VERSION" ] || die "failed to read required Go version from go.mod"

  if command -v go >/dev/null 2>&1; then
    if resolved=$(resolve_existing_go "$(command -v go)"); then
      printf '%s\n' "$resolved"
      exit 0
    fi

    found_version=$(go_version_at "$(command -v go)")
    log "Found Go $found_version on PATH, but Go $REQUIRED_GO_VERSION or newer is required."
  fi

  install_root="$(toolchain_root)/go/$REQUIRED_GO_VERSION"
  if resolved=$(resolve_existing_go "$install_root/bin/go"); then
    printf '%s\n' "$resolved"
    exit 0
  fi

  install_go "$install_root"
  printf '%s\n' "$install_root/bin/go"
}

main "$@"
