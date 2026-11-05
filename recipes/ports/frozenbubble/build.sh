#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
FILES=(
    "libtheoradec.so.1"
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
TTF_SOURCE_DIR="${SOURCE_DIR}"
GAME_CMAKE_ARGS=()

#### SDL_ttf ####
# The aarch64 builder (Ubuntu 20.04) ships SDL_ttf 2.0.15, which lacks TTF_SetFontSize (2.0.18)
# and TTF_SetFontWrappedAlign (2.20). Build the same series the x86_64 builder (Ubuntu 24.04) has.
if [[ ${ARCH} == "aarch64" ]]; then
  SDL2_TTF_PREFIX="${CDIR}/sdl2_ttf-prefix"
  git clone --depth 1 --branch release-2.22.0 https://github.com/libsdl-org/SDL_ttf.git || exit 1
  cmake -S SDL_ttf -B SDL_ttf/build -G Ninja -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="${SDL2_TTF_PREFIX}" -DCMAKE_INSTALL_LIBDIR=lib \
    -DBUILD_SHARED_LIBS=ON -DSDL2TTF_SAMPLES=OFF -DSDL2TTF_VENDORED=OFF -DSDL2TTF_HARFBUZZ=OFF || exit 1
  cmake --build SDL_ttf/build -j$(nproc) || exit 1
  cmake --install SDL_ttf/build || exit 1

  TTF_SOURCE_DIR="${SDL2_TTF_PREFIX}/lib"
  # The game includes <SDL2/SDL_ttf.h>, which would otherwise resolve to the old /usr/include header.
  # glibc 2.31 keeps pthread in a separate library and upstream doesn't link it (x86_64's glibc has it in libc).
  GAME_CMAKE_ARGS=(-DCMAKE_PREFIX_PATH="${SDL2_TTF_PREFIX}" -DCMAKE_CXX_FLAGS="-I${SDL2_TTF_PREFIX}/include"
    -DCMAKE_EXE_LINKER_FLAGS=-pthread)
fi

#### Frozen Bubble SDL2 ####
git clone --depth 1 https://github.com/dchau360/frozen-bubble-sdl2.git
cd frozen-bubble-sdl2

# PortMaster changes:
#  - FB_DATA_DIR env var for the asset folder (the default path is baked in at compile time)
#  - [Network] Host/Port in settings.ini pre-fill Net Game > Manual entry (written by the launcher)
#  - FB_SERVER_BIN env var / ./fb-server so in-game "Host a server" finds the bundled server
git apply --verbose "${CDIR}/recipes/ports/frozenbubble/frozenbubble-portmaster.patch" || exit 1

cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release "${GAME_CMAKE_ARGS[@]}"
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
    cp -L "${TTF_SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done
mkdir -p "${DEST_DIR}/fallback"
for file in "${FALLBACK_FILES[@]}"; do
    cp -L "${SOURCE_DIR}/${file}" "${DEST_DIR}/fallback/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
