#!/usr/bin/env bash
set -euo pipefail

# Builds a gomobile AAR for the cache bridge.
# Requires gomobile installed: go install golang.org/x/mobile/cmd/gomobile@latest
# and the Android SDK installed.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CACHEBRIDGE="$ROOT/core/cachebridge"
OUT="$ROOT/app/libs"

mkdir -p "$OUT"
cd "$CACHEBRIDGE"
gomobile bind -target=android -o "$OUT/cachebridge.aar" .
echo "AAR written to $OUT/cachebridge.aar"
