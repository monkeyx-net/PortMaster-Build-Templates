#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"

cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD}
mkdir -p dist
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp engine.config dist/
cp -r assets dist/

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
pwd
ls -lha
