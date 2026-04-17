#!/bin/sh
# Legnext CLI installer — curl -fsSL https://legnext.ai/install.sh | sh
set -e

REPO="legnext-ai/legnext-cli"
BINARY="legnext"
INSTALL_DIR="${LEGNEXT_INSTALL_DIR:-/usr/local/bin}"

# --- helpers ----------------------------------------------------------------

fail() { echo "Error: $1" >&2; exit 1; }

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) fail "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"v\([^"]*\)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"v\([^"]*\)".*/\1/'
  else
    fail "curl or wget required"
  fi
}

download() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
  else
    wget -qO "$2" "$1"
  fi
}

# --- main -------------------------------------------------------------------

OS="$(detect_os)"
ARCH="$(detect_arch)"
VERSION="${LEGNEXT_VERSION:-$(latest_version)}"

[ -z "$VERSION" ] && fail "could not determine latest version"

echo "Installing ${BINARY} v${VERSION} (${OS}/${ARCH})..."

if [ "$OS" = "windows" ]; then
  EXT="zip"
else
  EXT="tar.gz"
fi

ASSET="${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ASSET}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

download "$URL" "${TMP}/${ASSET}"

if [ "$EXT" = "zip" ]; then
  unzip -oq "${TMP}/${ASSET}" -d "$TMP"
else
  tar xzf "${TMP}/${ASSET}" -C "$TMP"
fi

# install binary
if [ -w "$INSTALL_DIR" ]; then
  cp "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"
else
  echo "Need sudo to install to ${INSTALL_DIR}"
  sudo cp "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  sudo chmod +x "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed ${BINARY} v${VERSION} to ${INSTALL_DIR}/${BINARY}"
"${INSTALL_DIR}/${BINARY}" --version
