#!/bin/bash
# xoji pre-task hook: Auto-refresh indexes before each Claude Code task
#
# This script runs automatically before each task in Claude Code,
# ensuring xoji indexes are always fresh without manual intervention.
#
# Usage: Claude Code calls this hook automatically if it exists

set -e

# Find project root (git repo or current directory)
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$PROJECT_ROOT"

# Path to xoji binary (adjust if stored elsewhere)
XOJI="${PROJECT_ROOT}/xoji"

# Check if binary exists
if [ ! -f "$XOJI" ]; then
    # Try to find xoji in PATH
    XOJI=$(command -v xoji || echo "")
    if [ -z "$XOJI" ]; then
        echo "Warning: xoji binary not found. Skipping index freshness check." >&2
        exit 0
    fi
fi

# Run freshness check; if stale, rebuild indexes
if ! "$XOJI" check 2>/dev/null; then
    echo "[xoji] Index stale, rebuilding..." >&2
    "$XOJI" index >/dev/null 2>&1
    echo "[xoji] Indexes refreshed" >&2
fi

exit 0
