#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="${PORT_FOLDER}/libs.${ARCH}"
FILES=(
    "libfluidsynth.so.2"
    "libinstpatch-1.0.so.2"
    "libjack.so.0"
    "libmad.so.0"
    "libmikmod.so.3"
    "libreadline.so.8"
    "libSDL_mixer-1.2.so.0"
    "libSDL-1.2.so.0"
    "libtinfo.so.6"
)


if [ [${ARCH} == "aarch64" ]] || [[ "${ARCH}" == "20.04.aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu/"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu/"
fi



echo ${PORT_FOLDER}
echo ${PORT_BUILD}
cd "new_ports/build/${PORT_FOLDER}"
${PORT_BUILD}
mkdir -p dist
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp -r data/ dist/
ls -lha ${SOURCE_DIR}

# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/"
done



tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
pwd
ls -lha
