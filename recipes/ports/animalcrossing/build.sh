#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"


cd "new_ports/${PORT_FOLDER}/source"
python3 confgiure.py
${PORT_BUILD}
mkdir -p dist/shaders
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp -r ../pc/shaders/* dist/shaders/

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
