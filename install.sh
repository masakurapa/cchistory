#!/usr/bin/env sh

set -eu

REPO="masakurapa/cchistory"
BIN_NAME="cchistory"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS="$(uname -s)"
ARCH="$(uname -m)"

if [ "$OS" != "Darwin" ]; then
  echo "Error: only macOS is supported."
  exit 1
fi

case "$ARCH" in
  arm64|aarch64)
    ASSET="${BIN_NAME}_darwin_arm64"
    ;;
  x86_64)
    ASSET="${BIN_NAME}_darwin_amd64"
    ;;
  *)
    echo "Error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${ASSET}..."

curl -fsSL \
  "https://github.com/${REPO}/releases/latest/download/${ASSET}" \
  -o "${TMP_DIR}/${BIN_NAME}"

chmod +x "${TMP_DIR}/${BIN_NAME}"

if [ -w "${INSTALL_DIR}" ]; then
  install -m 755 "${TMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
else
  sudo install -m 755 "${TMP_DIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
fi

# Ignore if the binary is signed/notarized or xattr is unavailable.
if command -v xattr >/dev/null 2>&1; then
  xattr -rd com.apple.quarantine "${INSTALL_DIR}/${BIN_NAME}" 2>/dev/null || true
fi

echo
echo "✅ ${BIN_NAME} installed successfully!"
echo
