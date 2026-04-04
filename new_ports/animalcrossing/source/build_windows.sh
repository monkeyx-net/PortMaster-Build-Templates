#!/bin/bash
# build_windows.sh - Cross-compile the Animal Crossing PC port for Windows
#
# Prerequisites: sudo apt install gcc-mingw-w64 g++-mingw-w64 cmake libsdl2-dev
#                (or equivalent for your distro)
#
# Usage:
#   1. ./build_windows.sh
#   2. Place your disc image (.ciso/.iso/.gcm) in pc/build_win/bin/rom/
#   3. Run: pc/build_win/bin/AnimalCrossing.exe (via Wine or on Windows)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/pc/build_win"
BIN_DIR="$BUILD_DIR/bin"

mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

if [ ! -f Makefile ]; then
    echo "=== Configuring CMake (Windows MinGW-w64 cross-compile) ==="
    cmake .. -G "Unix Makefiles" \
        -DCMAKE_TOOLCHAIN_FILE="../cmake/Toolchain-mingw-w64.cmake" \
        -DCMAKE_BUILD_TYPE=Release
fi

echo "=== Building ==="
make -j$(nproc)

# --- Create runtime directories ---
mkdir -p "$BIN_DIR/rom"
mkdir -p "$BIN_DIR/texture_pack"
mkdir -p "$BIN_DIR/save"

echo ""
echo "=== Build complete! ==="
echo ""
echo "Place your Animal Crossing disc image (.ciso/.iso/.gcm) in:"
echo "  pc/build_win/bin/rom/"
echo ""
echo "Run: pc/build_win/bin/AnimalCrossing.exe"
echo "  (Use Wine on Linux, or copy to Windows)"