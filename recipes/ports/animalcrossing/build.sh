#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"


cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD}
mkdir -p dist/shaders
cp "bin/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp -r bin/shaders/* dist/shaders/
ls -lha
ls -lha bin

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
