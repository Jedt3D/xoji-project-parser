#!/bin/bash
# Bump version number and build all binaries
# Usage: ./bump-version.sh [new-version]
#        ./bump-version.sh                    (interactive prompt)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$SCRIPT_DIR/src"
VERSION_FILE="$SRC_DIR/version.go"

# Current version
CURRENT_VERSION=$(grep 'const Version = ' "$VERSION_FILE" | sed 's/.*"\([^"]*\)".*/\1/')

echo "Current version: v$CURRENT_VERSION"
echo ""

# Get new version
if [ -z "$1" ]; then
    echo "Enter new version (e.g., 1.3.0):"
    read -r NEW_VERSION
    if [ -z "$NEW_VERSION" ]; then
        echo "Error: Version required"
        exit 1
    fi
else
    NEW_VERSION="$1"
fi

# Validate version format (basic check: should contain dots)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use semantic versioning (e.g., 1.3.0)"
    exit 1
fi

echo ""
echo "Updating version: v$CURRENT_VERSION → v$NEW_VERSION"
echo ""

# Update src/version.go
sed -i '' "s/const Version = \"[^\"]*\"/const Version = \"$NEW_VERSION\"/" "$VERSION_FILE"
echo "✓ Updated src/version.go"

# Build all binaries
echo ""
echo "Building all binaries..."
"$SCRIPT_DIR/build.sh"

# Offer to create git commit
echo ""
echo "Version bump complete!"
echo ""
echo "Next steps (optional):"
echo "  git add src/version.go"
echo "  git commit -m \"Bump version to v$NEW_VERSION\""
echo "  git tag v$NEW_VERSION"
echo "  git push origin main --tags"
echo ""
