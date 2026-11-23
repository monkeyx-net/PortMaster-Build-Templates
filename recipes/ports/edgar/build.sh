#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"

cd "new_ports/${PORT_FOLDER}/source"
# Scale the fixed 640x480 frame to any screen size (e.g. 720x720) instead of drawing it 1:1 in the corner
grep -q SDL_RenderSetLogicalSize src/init.c || \
  sed -i '/game.renderer = SDL_CreateRenderer/a\	SDL_RenderSetLogicalSize(game.renderer, SCREEN_WIDTH, SCREEN_HEIGHT);' src/init.c
${PORT_BUILD}
mkdir -p dist/libs.${ARCH}
cp "${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
if [[ ${ARCH} ==  "x86_64" ]]; then
  make buildpak -j$(nproc)
  cp "${PORT_EXE}".pak "dist/"
fi

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
