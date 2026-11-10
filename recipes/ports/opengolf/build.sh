#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# The Love2D launcher needs libtheoradec, which AmberELEC (aarch64 only) lacks
FILES=()
[[ ${ARCH} == "aarch64" ]] && FILES=(
  "libtheoradec.so.1"
)

cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD} || exit 1
cmake --build cmake-build -j$(nproc) || exit 1
mkdir -p dist
cp "cmake-build/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

mkdir -p "dist/libs.${ARCH}"
for file in "${FILES[@]}"; do
    path=$(ldconfig -p | awk -v l="$file" '$1 == l { print $NF; exit }')
    cp -L "$path" "dist/libs.${ARCH}/" 2>/dev/null || echo "Warning: ${file} not found"
done

# Game data goes in one archive that the launcher extracts on first run (hundreds of small
# files install slowly from the zip). Sorted, fixed mtime/owner and gzip -n, so both arch
# builds write the same assets.tar.gz and an unchanged build does not change the file.
rm -rf dist/data
tar --sort=name --mtime='2020-01-01 00:00Z' --owner=0 --group=0 --numeric-owner -cf - data \
    | gzip -9n > dist/assets.tar.gz

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
