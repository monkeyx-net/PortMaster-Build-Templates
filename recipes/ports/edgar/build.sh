#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"

cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD}
mkdir -p dist/libs.${ARCH}
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
if [[ ${ARCH} ==  "aarch64" ]]; then
  make buildpak -j$(nproc)
  cp "${PORT_EXE}".pak "dist/"
fi

tar -czf - -C dist . | split -b 75m - "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz."
