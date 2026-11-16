#!/bin/bash

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# The Love2D settings launcher needs libtheoradec on AmberELEC
FILES=()
[[ ${ARCH} == "aarch64" ]] && FILES+=(
  "libtheoradec.so.1"
)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"
SOURCE_DIR="/usr/lib/${ARCH}-linux-gnu"

RECIPE="${CDIR}/recipes/ports/${PORT_FOLDER}"
WORK="${CDIR}/new_ports/${PORT_FOLDER}/source/_work"
SHIM="${WORK}/SDL"
SHIM_COMMIT=6057d79baf8321bf190479a699655f06cc2a962f

# SDL3 is bmdhacks' sdl2-backend shim, which runs on the CFW's own SDL2. A real SDL3 loads
# libEGL itself, and on ArkOS that is glvnd/Mesa next to libMali's libgbm/libGLESv2, so the window
# fails. The patch (from PortMaster-New railroadrampage) fixes absolute mouse motion and a segfault
# on exit.
mkdir -p "$WORK"
[ -d "${SHIM}/.git" ] || git clone -b sdl2-backend https://github.com/bmdhacks/SDL.git "$SHIM" || exit 1
git -C "$SHIM" checkout -q "$SHIM_COMMIT" || exit 1
git -C "$SHIM" apply --reverse --check "${RECIPE}/sdl3-sdl2-backend-fixes.patch" 2>/dev/null \
  || git -C "$SHIM" apply "${RECIPE}/sdl3-sdl2-backend-fixes.patch" || exit 1

cd "new_ports/${PORT_FOLDER}/source"
# iniparser, miniaudio, libsmacker and libweaver are fetched at configure time; SDL3 is the shim,
# built shared (lego1 and isle both link it). Its GPU API is off: it links SPIRV-Cross and a static
# libstdc++, and isle doesn't use it.
# Not built: the OpenGL 1 renderer (links desktop libGL/GLX, missing on most CFWs), the SDL GPU
# renderer, the Qt config tool and libwebsockets multiplayer.
cmake -S . -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_DISABLE_FIND_PACKAGE_OpenGL=ON \
  -DISLE_RENDERER_SDLGPU=OFF \
  -DISLE_BUILD_CONFIG=OFF \
  -DISLE_USE_LWS=OFF \
  -DFETCHCONTENT_SOURCE_DIR_SDL3="$SHIM" \
  -DSDL_SDL2_BACKEND=ON -DSDL_GPU=OFF -DSDL_RENDER_GPU=OFF \
  -DSDL_X11=OFF -DSDL_WAYLAND=OFF -DSDL_KMSDRM=OFF -DSDL_VULKAN=OFF \
  -DSDL_PIPEWIRE=OFF -DSDL_PULSEAUDIO=OFF -DSDL_ALSA=OFF -DSDL_SNDIO=OFF -DSDL_OSS=OFF -DSDL_JACK=OFF \
  -DSDL_UNIX_CONSOLE_BUILD=ON || exit 1
grep -q '#define SDL_VIDEO_DRIVER_SDL2 1' build/_deps/sdl3-build/include-config-release/build_config/SDL_build_config.h \
  || { echo "Error: SDL3 was configured without the SDL2 backend"; exit 1; }
cmake --build build -j$(nproc) || exit 1

mkdir -p "${DEST_DIR}"
cp "build/${PORT_EXE}" "${CDIR}/dist/${PORT_EXE}.${ARCH}"
cp build/liblego1.so build/_deps/sdl3-build/libSDL3.so.0 "${DEST_DIR}/"
strip "${CDIR}/dist/${PORT_EXE}.${ARCH}" "${DEST_DIR}/liblego1.so" "${DEST_DIR}/libSDL3.so.0" || true
for file in "${FILES[@]}"; do
  cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" || echo "Warning: ${file} not found"
done
ls -lha "${CDIR}/dist" "${DEST_DIR}"

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C "${CDIR}/dist" .
