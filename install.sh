#!/bin/bash
set -e

REPO="d3vw/nft-ui"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="nft-ui"
BETA_MODE=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --beta)
            BETA_MODE=true
            shift
            ;;
    esac
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Check root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root: curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | sudo bash"
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       error "Unsupported architecture: $ARCH" ;;
esac

info "Detected architecture: $ARCH"

# Get latest release
if [ "$BETA_MODE" = true ]; then
    info "Fetching latest beta/pre-release..."
    LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases" | grep '"tag_name"' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')
else
    info "Fetching latest stable release..."
    LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi

if [ -z "$LATEST" ]; then
    error "Failed to get latest release"
fi

info "Latest version: $LATEST"

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY_NAME}-linux-${ARCH}"
info "Downloading from: $DOWNLOAD_URL"

curl -fsSL -o "/tmp/${BINARY_NAME}" "$DOWNLOAD_URL" || error "Download failed"

# Install
info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify
if command -v $BINARY_NAME &> /dev/null; then
    info "Installation successful!"
    echo ""
    echo "Usage:"
    echo "  $BINARY_NAME                    # Run with default settings"
    echo "  $BINARY_NAME -h                 # Show help"
    echo ""
    echo "Configuration (environment variables):"
    echo "  NFT_UI_LISTEN_ADDR=:8080        # Listen address"
    echo "  NFT_UI_AUTH_USER=admin          # Basic auth username"
    echo "  NFT_UI_AUTH_PASSWORD=secret     # Basic auth password"
    echo "  NFT_UI_READ_ONLY=false          # Read-only mode"
    echo ""
else
    error "Installation failed"
fi
