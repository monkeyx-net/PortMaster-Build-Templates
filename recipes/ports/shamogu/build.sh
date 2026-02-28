#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="${PORT_FOLDER}/libs.${ARCH}"

if [[ ${ARCH} == "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu/"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu/"
else
  echo "Have you chosen your ${ARCH} yet!"
fi

echo ${PORT_FOLDER}
echo ${PORT_BUILD}
cd "new_ports/build/${PORT_FOLDER}/source"
${PORT_BUILD}
mkdir -p dist
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
pwd
ls -lha
