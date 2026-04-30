#!/bin/bash

set -e

get_os() {
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        MINGW*|MSYS*|CYGWIN*) OS=windows;;
        *)          echo "Unsupported operating system"; exit 1;;
    esac
    echo "$OS"
}

get_arch() {
    case "$(uname -m)" in
        x86_64) ARCH=amd64;;
        arm64)  ARCH=arm64;;
        *)      echo "Unsupported architecture"; exit 1;;
    esac
    echo "$ARCH"
}

get_latest_version() {
    curl -s "https://api.github.com/repos/$OWNER/$NAME/releases/latest" | grep -o '"tag_name": ".*"' | sed 's/"tag_name": "//;s/"//'
}

OWNER="flowexec"
NAME="flow"
BINARY="flow"

OS=$(get_os)
ARCH=$(get_arch)
if [ -z "$VERSION" ]; then
    VERSION=$(get_latest_version)
fi

if [ "$OS" = "windows" ]; then
    EXT="zip"
    BINARY_NAME="${BINARY}.exe"
else
    EXT="tar.gz"
    BINARY_NAME="${BINARY}"
fi

DOWNLOAD_URL="https://github.com/${OWNER}/${NAME}/releases/download/${VERSION}/${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
TMP_DIR=$(mktemp -d)
DOWNLOAD_PATH="${TMP_DIR}/${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"

echo "Downloading $BINARY $VERSION for $OS/$ARCH..."
curl -fsSL "$DOWNLOAD_URL" -o "$DOWNLOAD_PATH"
if [ $? -ne 0 ]; then
    echo "Failed to download $DOWNLOAD_URL"
    exit 1
fi

if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
    echo "Installing $BINARY $VERSION to $INSTALL_DIR..."
    unzip -o -q "$DOWNLOAD_PATH" -d "$TMP_DIR"
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    INSTALL_DIR="/usr/local/bin"
    echo "Installing $BINARY $VERSION to $INSTALL_DIR..."
    tar -xzf "$DOWNLOAD_PATH" -C "$TMP_DIR"
    chmod +x "$TMP_DIR/$BINARY_NAME"
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

echo "$BINARY was installed successfully to $INSTALL_DIR/$BINARY_NAME"
if command -v $BINARY --version >/dev/null; then
    echo "Run '$BINARY --help' to get started"
else
    echo "Manually add the directory to your \$HOME/.bash_profile (or similar)"
    echo "  export PATH=$INSTALL_DIR:\$PATH"
fi

exit 0
