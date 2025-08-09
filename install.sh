#!/bin/bash

# SnapShell Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/saswatsam786/snapshell/main/install.sh | bash

set -e

REPO="saswatsam786/snapshell"
BINARY_NAME="snapshell"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Installing SnapShell...${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux)
        BINARY_NAME="snapshell-linux-amd64"
        ;;
    *)
        echo -e "${BLUE}🔧 To build from source:${NC}"
        echo ""
        echo -e "${GREEN}# Clone repository (requires access)${NC}"
        echo -e "git clone https://github.com/$REPO"
        echo -e "cd snapshell"
        echo ""
        echo -e "${GREEN}# Install dependencies${NC}"
        if [ "$OS" = "darwin" ]; then
            echo -e "brew install opencv pkg-config go"
        else
            echo -e "sudo apt-get install libopencv-dev libopencv-contrib-dev pkg-config golang-go"
        fi
        echo ""
        echo -e "${GREEN}# Build${NC}"
        echo -e "go build -o snapshell cmd/main.go"
        echo ""
        echo -e "${BLUE}For public distribution, the repository needs to be made public.${NC}"
        exit 1
        ;;
esac

# Get latest release
echo -e "${YELLOW}📡 Fetching latest release...${NC}"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${RED}❌ Failed to get latest release${NC}"
    exit 1
fi

echo -e "${GREEN}📦 Latest version: $LATEST_RELEASE${NC}"

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$BINARY_NAME"

# Create temp directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download binary
echo -e "${YELLOW}⬇️  Downloading $BINARY_NAME...${NC}"
curl -sSL "$DOWNLOAD_URL" -o "$BINARY_NAME"

# Make executable
chmod +x "$BINARY_NAME"

# Install to system
echo -e "${YELLOW}📦 Installing to $INSTALL_DIR...${NC}"
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/snapshell"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/snapshell"
fi

# Cleanup
cd ..
rm -rf "$TMP_DIR"

echo -e "${GREEN}✅ SnapShell installed successfully!${NC}"
echo
echo -e "${BLUE}🎥 Quick Start:${NC}"
echo -e "  ${YELLOW}# Start video sharing session${NC}"
echo -e "  snapshell -signaled-o --room demo123 --server https://snapshell.onrender.com"
echo -e "  snapshell -signaled-a --room demo123 --server https://snapshell.onrender.com"
echo
echo -e "${BLUE}📖 More info: https://github.com/$REPO${NC}"
