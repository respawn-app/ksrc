#!/usr/bin/env sh
set -eu

REPO="${KSRC_REPO:-respawn-app/ksrc}"
VERSION="${KSRC_VERSION:-${VERSION:-}}"
PREFIX="${KSRC_PREFIX:-}"
RELEASE_BASE="${KSRC_RELEASE_BASE:-https://github.com/${REPO}/releases/download}"

usage() {
  cat <<EOF
Usage: install.sh [--version vX.Y.Z] [--prefix /path]

Options:
  --version  Release tag to install (defaults to latest)
  --prefix   Install prefix (defaults to /usr/local or ~/.local)

Environment:
  KSRC_VERSION       Override version
  KSRC_PREFIX        Override prefix
  KSRC_REPO          Override repo (default: respawn-app/ksrc)
  KSRC_RELEASE_BASE  Override release base URL
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --prefix)
      PREFIX="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | awk -F'"' '/"tag_name":/ {print $4; exit}')"
fi
if [ -z "$VERSION" ]; then
  echo "Failed to resolve latest version." >&2
  exit 1
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  mingw*|msys*|cygwin*) os="windows" ;;
  *)
    echo "Unsupported OS: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

ver="${VERSION#v}"
base_name="ksrc_${ver}_${os}_${arch}"
if [ "$os" = "windows" ]; then
  archive="${base_name}.zip"
  bin_name="${base_name}.exe"
else
  archive="${base_name}.tar.gz"
  bin_name="${base_name}"
fi

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

url="${RELEASE_BASE}/${VERSION}/${archive}"
curl -fsSL "$url" -o "$tmpdir/$archive"

if [ "$os" = "windows" ]; then
  if ! command -v unzip >/dev/null 2>&1; then
    echo "unzip is required to install on Windows." >&2
    exit 1
  fi
  unzip -q "$tmpdir/$archive" -d "$tmpdir"
else
  tar -xzf "$tmpdir/$archive" -C "$tmpdir"
fi

if [ -z "$PREFIX" ]; then
  if [ -w /usr/local/bin ]; then
    PREFIX="/usr/local"
  else
    PREFIX="$HOME/.local"
  fi
fi

bin_dir="$PREFIX"
case "$bin_dir" in
  */bin) ;;
  *) bin_dir="${bin_dir%/}/bin" ;;
esac

mkdir -p "$bin_dir"
install -m 755 "$tmpdir/$bin_name" "$bin_dir/ksrc"

echo "Installed ksrc to $bin_dir/ksrc"
if ! echo "$PATH" | tr ':' '\n' | grep -q "^${bin_dir}$"; then
  echo "Note: $bin_dir is not on PATH. Add it to your shell profile."
fi
