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

cd "new_ports/${PORT_FOLDER}/source"
# SDL3, iniparser, miniaudio, libsmacker and libweaver are fetched at configure time. CFWs ship
# only SDL2, so SDL3 is bundled (built shared, as lego1 and isle both link it).
# Not built: the OpenGL 1 renderer (links desktop libGL/GLX, missing on most CFWs), the SDL GPU
# renderer (needs Vulkan), the Qt config tool and libwebsockets multiplayer. XTEST is not in the
# 20.04 builder (and SDL3 only needs it for X11 input faking).
cmake -S . -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_DISABLE_FIND_PACKAGE_OpenGL=ON \
  -DISLE_RENDERER_SDLGPU=OFF \
  -DISLE_BUILD_CONFIG=OFF \
  -DISLE_USE_LWS=OFF \
  -DSDL_X11_XTEST=OFF || exit 1
grep -q '#define SDL_VIDEO_DRIVER_KMSDRM 1' build/_deps/sdl3-build/include-config-release/build_config/SDL_build_config.h \
  || { echo "Error: SDL3 was configured without KMSDRM"; exit 1; }
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
