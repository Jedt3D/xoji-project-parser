#!/bin/bash
# xoji build script for macOS and Linux
# Builds cross-platform binaries to dist/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$SCRIPT_DIR/src"
DIST_DIR="$SCRIPT_DIR/dist"
VERSION="1.0.0"

echo "Building xoji v$VERSION..."
mkdir -p "$DIST_DIR"

cd "$SRC_DIR"

# macOS ARM64 (Apple Silicon)
echo "  → macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o "$DIST_DIR/xoji-mac-arm64" .

# macOS x86-64 (Intel)
echo "  → macOS x86-64..."
GOOS=darwin GOARCH=amd64 go build -o "$DIST_DIR/xoji-mac-amd64" .

# Linux x86-64
echo "  → Linux x86-64..."
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/xoji-linux" .

# Windows x86-64
echo "  → Windows x86-64..."
GOOS=windows GOARCH=amd64 go build -o "$DIST_DIR/xoji-windows.exe" .

# Also build for current platform in root (for immediate use)
echo "  → Native build..."
go build -o "$SCRIPT_DIR/xoji" .

echo "✓ Build complete!"
echo ""
echo "Binaries in dist/:"
ls -lh "$DIST_DIR" | tail -n +2 | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo "Quick start:"
echo "  ./xoji setup ../my_xojo_project"
echo "  ./xoji index ../my_xojo_project"
