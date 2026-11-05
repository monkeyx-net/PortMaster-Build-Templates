#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
CDIR=$(pwd)

FILES=(
  "libtheoradec.so.1"
)

cd "new_ports/${PORT_FOLDER}/source"

# The fork's build.sh builds Allegro 5 (SDL2 + GLES2 backend) and Lix, and
# assembles the port in $OUTDIR/lix. Drop its debug-only libSegFault copy
# (not in glibc 2.36+, so it fails on the x86_64 image), and use Allegro
# 5.2.10's option names so the addons don't link libwebp/libdumb.
sed -i -e '/libSegFault/d' \
    -e 's/-DWANT_WEBP=OFF/-DWANT_IMAGE_WEBP=OFF/' \
    -e 's/-DWANT_MODAUDIO=OFF/-DWANT_DUMB=OFF -DWANT_OPENMPT=OFF/' "${PORT_BUILD}"
OUTDIR="$PWD/_out" bash "${PORT_BUILD}"

# Launcher, port.json, gptk and README live in the port folder of this repo
mkdir -p "${CDIR}/dist"
cd _out/lix
mv "lix.${ARCH}" "${CDIR}/dist/${PORT_EXE}.${ARCH}"
strip "${CDIR}/dist/${PORT_EXE}.${ARCH}" || true
rm -f port.json lix.gptk README.md allegro5.cfg
cp -r . "${CDIR}/dist/"

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C "${CDIR}/dist" .
