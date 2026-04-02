#!/usr/bin/env bash
set -e

REPO="Harsh-2002/Zeno"
BINARY="zeno"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)              echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
echo "→ Detected: ${OS}/${ARCH}"

# Get latest release tag
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$TAG" ]; then
  echo "Error: Could not find latest release"
  exit 1
fi
echo "→ Latest release: ${TAG}"

# Download
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
TMP=$(mktemp)
echo "→ Downloading ${ASSET}..."
curl -fsSL -o "$TMP" "$URL"
chmod +x "$TMP"

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

mv "$TMP" "${INSTALL_DIR}/${BINARY}"
echo "→ Installed to ${INSTALL_DIR}/${BINARY}"

# Verify
if command -v "$BINARY" &>/dev/null; then
  echo "→ Done! Run: zeno"
else
  echo "→ Done! Add ${INSTALL_DIR} to your PATH, then run: zeno"
fi
