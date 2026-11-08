#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"

cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD} || exit 1
cmake --build cmake-build -j$(nproc) || exit 1
mkdir -p dist
cp "cmake-build/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

# Game data (both arch builds produce the same files)
rm -rf dist/data
cp -r data dist/data

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
