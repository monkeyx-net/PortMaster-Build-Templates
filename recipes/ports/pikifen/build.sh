#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
CDIR=$(pwd)
RECIPE="${CDIR}/recipes/ports/${PORT_FOLDER}"
DIST="${CDIR}/dist"
LIBS="${DIST}/libs.${ARCH}"
WORK="${CDIR}/new_ports/${PORT_FOLDER}/source/_work"
PREFIX="${WORK}/prefix"
ALLEGRO_VERSION=5.2.10.0

# System libraries devices may lack, bundled in libs.<arch>
FILES=(
  "libfreetype.so.6"
  "libpng16.so.16"
  "libjpeg.so.8"
  "libvorbis.so.0"
  "libvorbisfile.so.3"
  "libogg.so.0"
)

mkdir -p "$WORK"

# Allegro 5 with the SDL2 backend and GLES2 renderer (devices have no X11 or desktop GL)
if [ ! -f "${PREFIX}/lib/liballegro.so" ]; then
    rm -rf "${WORK}/allegro5"
    git clone --depth 1 -b "$ALLEGRO_VERSION" https://github.com/liballeg/allegro5.git "${WORK}/allegro5"
    git -C "${WORK}/allegro5" apply "${RECIPE}/allegro-sdl-joystick-without-driver.patch"
    cmake -S "${WORK}/allegro5" -B "${WORK}/allegro5/build" \
        -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="$PREFIX" \
        -DALLEGRO_SDL=ON -DGL_BUILD_TYPE=gles2+ -DWANT_X11=OFF \
        -DWANT_DOCS=OFF -DWANT_EXAMPLES=OFF -DWANT_TESTS=OFF -DWANT_DEMO=OFF \
        -DWANT_VIDEO=OFF -DWANT_PHYSFS=OFF -DWANT_NATIVE_DIALOG=OFF \
        -DWANT_FLAC=OFF -DWANT_DUMB=OFF -DWANT_OPENMPT=OFF -DWANT_OPUS=OFF -DWANT_MP3=OFF \
        -DWANT_IMAGE_WEBP=OFF -DWANT_OPENAL=OFF -DWANT_PULSEAUDIO=OFF \
        -DWANT_ALSA=OFF -DWANT_OSS=OFF
    cmake --build "${WORK}/allegro5/build" -j"$(nproc)"
    cmake --install "${WORK}/allegro5/build"
    # The native dialog addon needs GTK. The patch stubs the dialog calls,
    # the header only provides the constants.
    cp "${WORK}/allegro5/addons/native_dialog/allegro5/allegro_native_dialog.h" "${PREFIX}/include/allegro5/"
fi

# Pikifen, patched for GLSL ES shaders and without native dialogs
cd "new_ports/${PORT_FOLDER}/source"
# (-N and the grep allow re-running on an already patched source)
patch -p1 -N -r - < "${RECIPE}/pikifen-portmaster.patch" \
    || grep -q "#version 100" source/source/core/shaders_source.cpp
export PKG_CONFIG_PATH="${PREFIX}/lib/pkgconfig"
# Ubuntu 20.04's gcc 9 is too old for Pikifen
${PORT_BUILD} CXX=clang++-18 -j"$(nproc)"

rm -rf "$DIST"
mkdir -p "$LIBS"
cp "${PORT_EXE}" "${DIST}/${PORT_EXE}.${ARCH}"
strip "${DIST}/${PORT_EXE}.${ARCH}" || true
cp -r game_data "$DIST/"

# No symlinks: they break on FAT/exFAT SD cards
cp -L "${PREFIX}"/lib/liballegro*.so.5.2 "$LIBS/"
for file in "${FILES[@]}"; do
    path=$(ldconfig -p | awk -v l="$file" '$1 == l { print $NF; exit }')
    cp -L "$path" "$LIBS/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C "$DIST" .
