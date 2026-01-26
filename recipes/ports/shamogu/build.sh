#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"

ls -lha /lib/aarch64-linux-gnu/
ls -lha /usr/lib/aarch64-linux-gnu/

echo ${PORT_FOLDER}
echo ${PORT_BUILD}
cd "new_ports/build/${PORT_FOLDER}"
${PORT_BUILD}
mkdir -p dist
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
pwd
ls -lha
