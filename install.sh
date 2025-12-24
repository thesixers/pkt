#!/bin/bash
# PKT Universal Package Manager - Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/YOUR_USERNAME/pkt/main/install.sh | bash

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         PKT - Universal Package Manager                   ║"
echo "║                  Installation                              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS names
case "$OS" in
    linux)
        BINARY="pkt-linux-${ARCH}"
        ;;
    darwin)
        BINARY="pkt-macos-${ARCH}"
        ;;
    mingw*|msys*|cygwin*)
        BINARY="pkt-windows-x64.exe"
        ;;
    *)
        echo "❌ Unsupported operating system: $OS"
        exit 1
        ;;
esac

echo "📦 Detected: $OS ($ARCH)"
echo "📥 Downloading PKT..."

# Download latest release
GITHUB_REPO="thesixers/pkt"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${BINARY}"

if command -v curl &> /dev/null; then
    curl -L "$DOWNLOAD_URL" -o pkt
elif command -v wget &> /dev/null; then
    wget "$DOWNLOAD_URL" -O pkt
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

# Make executable
chmod +x pkt

# Install to system
INSTALL_DIR="/usr/local/bin"

if [ -w "$INSTALL_DIR" ]; then
    mv pkt "$INSTALL_DIR/pkt"
else
    echo "🔐 Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv pkt "$INSTALL_DIR/pkt"
fi

echo ""
echo "✅ PKT installed successfully!"
echo ""

# Verify installation
if command -v pkt &> /dev/null; then
    pkt version
    echo ""
    echo "🎉 You can now use 'pkt' from anywhere!"
    echo ""
    echo "Quick start:"
    echo "  pkt help                    # Show all commands"
    echo "  pkt create --language node  # Create a project"
    echo "  pkt add react               # Add a dependency"
    echo ""
    echo "Documentation: https://github.com/${GITHUB_REPO}"
else
    echo "⚠️  Installation complete, but 'pkt' not found in PATH"
    echo "   You may need to restart your terminal or add $INSTALL_DIR to PATH"
fi
