#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
FILES=(
  "libSDL2_ttf-2.0.so.0"
)
# fb-server needs glib; most firmwares ship it, so these are only used as a
# fallback (searched after the system libraries) to avoid clashing with the
# system's own glib/gobject.
FALLBACK_FILES=(
  "libglib-2.0.so.0"
  "libpcre.so.3"
  "libpcre2-8.so.0"
)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### Frozen Bubble SDL2 ####
git clone --depth 1 https://github.com/dchau360/frozen-bubble-sdl2.git
cd frozen-bubble-sdl2

# PortMaster changes:
#  - FB_DATA_DIR env var for the asset folder (the default path is baked in at compile time)
#  - [Network] Host/Port in settings.ini pre-fill Net Game > Manual entry (written by the launcher)
#  - FB_SERVER_BIN env var / ./fb-server so in-game "Host a server" finds the bundled server
git apply --verbose "${CDIR}/recipes/ports/frozenbubble/frozenbubble-portmaster.patch" || exit 1

cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release
cmake --build build -j$(nproc)
cd ..

if [[ ! -f "frozen-bubble-sdl2/build/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
if [[ ! -f "frozen-bubble-sdl2/build/server/fb-server" ]]; then
  echo "Error: fb-server was not built"
  exit 1
fi

mkdir -p "${CDIR}/dist"
strip "frozen-bubble-sdl2/build/${PORT_EXE}" || true
strip "frozen-bubble-sdl2/build/server/fb-server" || true
cp "frozen-bubble-sdl2/build/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "frozen-bubble-sdl2/build/server/fb-server" "dist/fb-server.${ARCH}"

# Game assets (both arch builds produce the same files)
rm -rf "${CDIR}/dist/share"
cp -r frozen-bubble-sdl2/share "${CDIR}/dist/share"
rm -rf "${CDIR}/dist/share/icons"
find "${CDIR}/dist/share" -name "*.md" -delete
ls -lha dist/

mkdir -p ${DEST_DIR}
for file in "${FILES[@]}"; do
    cp -L "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done
mkdir -p "${DEST_DIR}/fallback"
for file in "${FALLBACK_FILES[@]}"; do
    cp -L "${SOURCE_DIR}/${file}" "${DEST_DIR}/fallback/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
