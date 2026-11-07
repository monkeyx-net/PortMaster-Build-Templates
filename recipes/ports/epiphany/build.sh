#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"

cd "new_ports/${PORT_FOLDER}/source"

# Start fullscreen so the game fills the screen on KMSDRM
sed -i 's/m_fullscreen = false;/m_fullscreen = true;/' src/epiconfig.cpp

autoreconf -fi
./configure LDFLAGS="-L/usr/local/lib"
make -j$(nproc)

mkdir -p dist/data
cp "src/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp -r data/gfx data/sfx data/maps data/epiphany.scr dist/data/
rm -f dist/data/*/Makefile* dist/data/*/.cvsignore

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
