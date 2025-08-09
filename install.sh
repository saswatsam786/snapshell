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
    darwin)
        echo -e "${YELLOW}🍺 For macOS, we recommend using Homebrew:${NC}"
        echo ""
        echo -e "${GREEN}# Official Homebrew Tap (recommended)${NC}"
        echo -e "brew tap saswatsam786/snapshell"
        echo -e "brew install --HEAD snapshell"
        echo ""
        echo -e "${GREEN}# Alternative: Local formula${NC}"
        echo -e "git clone https://github.com/$REPO.git"
        echo -e "cd snapshell"
        echo -e "brew install --build-from-source --HEAD --formula ./Formula/snapshell.rb"
        echo ""
        echo -e "${BLUE}Homebrew automatically handles OpenCV and all dependencies!${NC}"
        exit 0
        ;;
    *)
        echo -e "${RED}❌ Currently only Linux x86_64 is supported via binary releases${NC}"
        echo -e "${YELLOW}Your system: $OS-$ARCH${NC}"
        echo ""
        echo -e "${BLUE}🍺 For macOS, use Homebrew:${NC}"
        echo -e "  brew install https://raw.githubusercontent.com/$REPO/main/Formula/snapshell.rb"
        echo ""
        echo -e "${BLUE}🔧 Or build from source:${NC}"
        echo -e "  git clone https://github.com/$REPO"
        echo -e "  cd snapshell"
        echo -e "  go build -o snapshell cmd/main.go"
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
