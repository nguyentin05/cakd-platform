#!/bin/bash
set -e

# CAKD CLI Installation Script
# https://github.com/nguyentin05/cakd-platform

OWNER="nguyentin05"
REPO="cakd-platform"
BIN_NAME="cakd"
INSTALL_DIR="/usr/local/bin"

echo "Looking for the latest release of $BIN_NAME..."

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_NAME="linux";;
    Darwin*)    OS_NAME="darwin";;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64*)    ARCH_NAME="amd64";;
    arm64*)     ARCH_NAME="arm64";;
    aarch64*)   ARCH_NAME="arm64";;
    *)          echo "Unsupported Architecture: ${ARCH}"; exit 1;;
esac

# Get latest release URL from GitHub API
LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/$OWNER/$REPO/releases/latest | grep "browser_download_url" | grep "${OS_NAME}_${ARCH_NAME}.tar.gz" | cut -d '"' -f 4)

if [ -z "$LATEST_RELEASE_URL" ]; then
    echo "Could not find a release for ${OS_NAME} ${ARCH_NAME}."
    echo "Please check https://github.com/$OWNER/$REPO/releases"
    exit 1
fi

echo "Downloading $LATEST_RELEASE_URL..."
curl -sL "$LATEST_RELEASE_URL" -o "/tmp/${BIN_NAME}.tar.gz"

echo "Extracting..."
tar -xzf "/tmp/${BIN_NAME}.tar.gz" -C /tmp

echo "Installing to $INSTALL_DIR..."
# Require sudo if not root
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
    echo "You might be prompted for your password to install to $INSTALL_DIR"
else
    SUDO=""
fi

$SUDO mv "/tmp/${BIN_NAME}" "$INSTALL_DIR/${BIN_NAME}"
$SUDO chmod +x "$INSTALL_DIR/${BIN_NAME}"

# Clean up
rm "/tmp/${BIN_NAME}.tar.gz"

echo ""
echo "====================================================="
echo "✅ $BIN_NAME installed successfully to $INSTALL_DIR!"
echo "Run '$BIN_NAME --help' to get started."
echo "====================================================="
