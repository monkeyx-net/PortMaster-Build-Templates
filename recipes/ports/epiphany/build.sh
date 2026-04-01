#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"

FILES=(
    "libtinfo.so.6"
    "libreadline.so.8"
    "libmikmod.so.3"
    "libmad.so.0"
    "libjack.so.0"
    "libinstpatch-1.0.so.2"
    "libfluidsynth.so.2"
    "libSDL_mixer-1.2.so.0"
    "libSDL-1.2.so.0"
    "libFLAC.so.8"
)

DEST_DIR="${CDIR}/dist/libs.${ARCH}"

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

cd "new_ports/${PORT_FOLDER}/source"
${PORT_BUILD}
mkdir -p dist/libs.${ARCH}
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
