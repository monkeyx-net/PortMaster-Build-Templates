#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
WORK="$(pwd)/new_ports/${PORT_FOLDER}/source/_work"
LUA_VERSION=5.4.7
LUA_SHA256=9fbf5e28ef86c69858f6d3d34eccc32e911c1a28b4120ff3e84aaa70cfbf1e30

# Lua is not in the builder images. Upstream links its own static Lua 5.4; liblua.a,
# libjpeg.a and libbz2.a go in one folder searched first, so they are linked statically
# (libbz2.so.1.0 is missing on non-Debian CFWs) and nothing has to be bundled.
mkdir -p "${WORK}/static"
LUA_SRC="${WORK}/lua-${LUA_VERSION}"
if [ ! -f "${LUA_SRC}/src/liblua.a" ]; then
    curl -fsSL -o "${WORK}/lua.tar.gz" "https://www.lua.org/ftp/lua-${LUA_VERSION}.tar.gz"
    echo "${LUA_SHA256}  ${WORK}/lua.tar.gz" | sha256sum -c -
    tar -xzf "${WORK}/lua.tar.gz" -C "$WORK"
    make -C "${LUA_SRC}/src" liblua.a SYSCFLAGS="-DLUA_USE_LINUX" -j"$(nproc)"
fi
ln -sf "${LUA_SRC}/src/liblua.a" "${WORK}/static/liblua.a"
for lib in jpeg bz2; do
    ln -sf "$(gcc -print-file-name=lib${lib}.a)" "${WORK}/static/lib${lib}.a"
done
# FindLua looks in $LUA_DIR; CMake adds $LDFLAGS to the link flags
export LUA_DIR="${LUA_SRC}/src" LDFLAGS="-L${WORK}/static"

cd "new_ports/${PORT_FOLDER}/source"
bash -c "${PORT_BUILD}"
rm -rf dist
mkdir -p dist
cp "build/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
