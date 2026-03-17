#!/usr/bin/env sh

set -eu

REPO="weiyong1024/ClawFleet"
BINARY="clawfleet"

log() {
  printf '==> %s\n' "$*" >&2
}

die() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Darwin) printf 'darwin\n' ;;
    Linux)  printf 'linux\n' ;;
    *)      die "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    arm64|aarch64) printf 'arm64\n' ;;
    x86_64|amd64)  printf 'amd64\n' ;;
    *)             die "unsupported architecture: $(uname -m)" ;;
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

  die "curl or wget is required"
}

sha256_verify() {
  file=$1
  expected=$2

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$file" | awk '{ print $1 }')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$file" | awk '{ print $1 }')
  else
    log "WARNING: no sha256sum or shasum found, skipping checksum verification"
    return 0
  fi

  if [ "$actual" != "$expected" ]; then
    die "checksum mismatch: expected $expected, got $actual"
  fi
}

latest_version() {
  url="https://api.github.com/repos/${REPO}/releases/latest"

  if command -v curl >/dev/null 2>&1; then
    tag=$(curl -fsSL "$url" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
  elif command -v wget >/dev/null 2>&1; then
    tag=$(wget -qO- "$url" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
  else
    die "curl or wget is required"
  fi

  [ -n "$tag" ] || die "failed to determine latest release"
  printf '%s\n' "$tag"
}

install_binary() {
  src=$1

  # Try /usr/local/bin first
  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    install_dir="/usr/local/bin"
  elif [ -w "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    install_dir="$HOME/.local/bin"
  else
    die "cannot find a writable install directory. Run with sudo or create ~/.local/bin"
  fi

  cp "$src" "$install_dir/$BINARY"
  chmod +x "$install_dir/$BINARY"

  log "Installed $BINARY to $install_dir/$BINARY"

  # Check if install_dir is in PATH
  case ":$PATH:" in
    *":$install_dir:"*) ;;
    *) log "Add \"$install_dir\" to your PATH to use $BINARY from anywhere." ;;
  esac
}

usage() {
  cat >&2 <<'EOF'
Usage: install.sh [--version <tag>]

Install ClawFleet from GitHub releases.

Options:
  --version <tag>   Install a specific version (e.g. v0.1.0). Default: latest.
EOF
}

main() {
  version=""

  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        [ $# -ge 2 ] || die "--version requires a value"
        version="$2"
        shift 2
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        die "unknown option: $1"
        ;;
    esac
  done

  os=$(detect_os)
  arch=$(detect_arch)

  if [ -z "$version" ]; then
    log "Fetching latest release..."
    version=$(latest_version)
  fi

  # Strip leading 'v' for archive name (GoReleaser uses version without 'v')
  ver_num="${version#v}"

  archive_name="${BINARY}_${ver_num}_${os}_${arch}.tar.gz"
  base_url="https://github.com/${REPO}/releases/download/${version}"

  tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/clawfleet-install.XXXXXX")
  trap 'rm -rf "$tmp_dir"' EXIT INT TERM HUP

  log "Downloading ${archive_name}..."
  download_file "${base_url}/${archive_name}" "$tmp_dir/${archive_name}"

  log "Downloading checksums..."
  download_file "${base_url}/checksums.txt" "$tmp_dir/checksums.txt"

  expected=$(grep "${archive_name}" "$tmp_dir/checksums.txt" | awk '{ print $1 }')
  [ -n "$expected" ] || die "checksum for ${archive_name} not found in checksums.txt"

  log "Verifying checksum..."
  sha256_verify "$tmp_dir/${archive_name}" "$expected"

  log "Extracting..."
  tar -C "$tmp_dir" -xzf "$tmp_dir/${archive_name}"

  install_binary "$tmp_dir/$BINARY"

  log "ClawFleet ${version} installed successfully!"
  log "Run '$BINARY version' to verify."
}

main "$@"
