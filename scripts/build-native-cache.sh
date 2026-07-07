#!/usr/bin/env bash
set -euo pipefail

# Builds the Go bbolt cache bridge for Android ABIs.
# Requires the Android NDK installed and ANDROID_NDK_HOME set.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CACHEBRIDGE="$ROOT/core/cachebridge"
OUT="$ROOT/app/src/main/jniLibs"

: "${ANDROID_NDK_HOME:?ANDROID_NDK_HOME must be set}"

export CGO_ENABLED=1

build_abi() {
    local abi=$1
    local cc=$2
    local goarch=$3
    local outdir="$OUT/$abi"
    mkdir -p "$outdir"
    echo "Building $abi..."
    CC="$cc" GOOS=android GOARCH="$goarch" go build \
        -buildmode=c-shared \
        -ldflags="-s -w" \
        -o "$outdir/libhermdroidcache.so" \
        "$CACHEBRIDGE/cmd/hermdroidcache"
}

build_abi arm64-v8a "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang" arm64
build_abi armeabi-v7a "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi21-clang" arm
build_abi x86_64 "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang" amd64
build_abi x86 "$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang" 386

echo "Native cache libraries built in $OUT"
