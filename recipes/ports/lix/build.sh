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
LDC_VERSION=1.39.0

# System libraries devices may lack, bundled in libs.<arch>
FILES=(
  "libenet.so.7"
  "libjpeg.so.8"
  "libvorbis.so.0"
  "libvorbisfile.so.3"
  "libogg.so.0"
  "libtheoradec.so.1"
)

mkdir -p "$WORK"

# LDC: the image's Ubuntu 20.04 ldc is too old for Lix
LDC_DIR="${WORK}/ldc2-${LDC_VERSION}-linux-${ARCH}"
[ -x "${LDC_DIR}/bin/ldc2" ] || curl -fsSL \
    "https://github.com/ldc-developers/ldc/releases/download/v${LDC_VERSION}/ldc2-${LDC_VERSION}-linux-${ARCH}.tar.xz" \
    | tar -xJ -C "$WORK"
export PATH="${LDC_DIR}/bin:$PATH"

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
fi

# Lix, patched for an integer-scaled mouse cursor (LIX_CURSOR_MIN_SCALE), no cursor
# jump on mouse-enter events (sent on button presses under Wayland) and
# more music buffering (Allegro's SDL driver reads 4096 frames at a time)
cd "new_ports/${PORT_FOLDER}/source"
# (-N and the grep allow re-running on an already patched source)
patch -p1 -N -r - < "${RECIPE}/lix-portmaster.patch" \
    || grep -q "4, 2048" src/hardware/music.d
export LIBRARY_PATH="${PREFIX}/lib" PKG_CONFIG_PATH="${PREFIX}/lib/pkgconfig"
${PORT_BUILD}

rm -rf "$DIST"
mkdir -p "$LIBS"
cp "bin/${PORT_EXE}" "${DIST}/${PORT_EXE}.${ARCH}"
strip "${DIST}/${PORT_EXE}.${ARCH}" || true
ASSETS="${WORK}/assets"
rm -rf "$ASSETS"
mkdir -p "$ASSETS"
cp -r data images levels "$ASSETS/"
rm -rf "${ASSETS}/data/desktop"
# Bigger, outlined mouse cursor for handheld screens
cp -r "${RECIPE}/data/." "${ASSETS}/data/"

# No symlinks: they break on FAT/exFAT SD cards
cp -L "${PREFIX}"/lib/liballegro*.so.5.2 "$LIBS/"
for file in "${FILES[@]}"; do
    path=$(ldconfig -p | awk -v l="$file" '$1 == l { print $NF; exit }')
    cp -L "$path" "$LIBS/" 2>/dev/null || echo "Warning: ${file} not found"
done
# Lix loads enet at runtime (Derelict), which looks for libenet.so
cp "$LIBS/libenet.so.7" "$LIBS/libenet.so"

curl -fsSL -o "${WORK}/lix-music.zip" https://www.lixgame.com/dow/lix-music.zip
unzip -q -o "${WORK}/lix-music.zip" -d "$ASSETS/"

# Game files go in one archive that the launcher extracts on first run (thousands of small
# files install slowly from the zip). Sorted, fixed mtime/owner and gzip -n, so both arch
# builds write the same assets.tar.gz and an unchanged build does not change the file.
tar --sort=name --mtime='2020-01-01 00:00Z' --owner=0 --group=0 --numeric-owner -C "$ASSETS" -cf - data images levels music \
    | gzip -9n > "${DIST}/assets.tar.gz"

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C "$DIST" .
