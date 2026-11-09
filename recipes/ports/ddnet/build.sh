#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
CDIR="${PWD}"
SRC="${CDIR}/new_ports/${PORT_FOLDER}/source"
DEPS="${SRC}/deps.${ARCH}"
LIBDIR="/usr/lib/${ARCH}-linux-gnu"

# DDNet needs Rust >= 1.85; the builder images do not ship Rust (HOME=/tmp in the build container)
export CARGO_HOME="${HOME}/.cargo"
export RUSTUP_HOME="${HOME}/.rustup"
if ! command -v cargo >/dev/null 2>&1; then
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
    | sh -s -- -y --profile minimal --default-toolchain stable --no-modify-path
fi
export PATH="${CARGO_HOME}/bin:${PATH}"

# DDNet is C++20 (std::bit_cast), which the Ubuntu 20.04 libstdc++ lacks: on aarch64 use clang
# with libc++ linked statically (also for the cxx crate's C++ that cargo builds)
CMAKE_ARGS=()
if [[ "${ARCH}" == "aarch64" ]]; then
  export CC=clang-18 CXX=clang++-18 CXXFLAGS="-stdlib=libc++" CXXSTDLIB=c++
  CMAKE_ARGS=(-DCMAKE_EXE_LINKER_FLAGS="-fuse-ld=lld -nostdlib++ -Wl,-Bstatic -lc++ -lc++abi -Wl,-Bdynamic")
fi

# curl, SQLite and WavPack are not on every CFW (and distro libcurl drags in many more libs):
# build them as static libraries. curl uses the builder's static OpenSSL and has no compiled-in
# CA bundle, so it falls back to OpenSSL's default paths / SSL_CERT_FILE (set by the launcher).
if [[ ! -f "${DEPS}/lib/libcurl.a" ]]; then
  mkdir -p "${DEPS}/src" && cd "${DEPS}/src"
  wget -q https://curl.se/download/curl-8.11.1.tar.xz
  wget -q https://www.sqlite.org/2024/sqlite-amalgamation-3470200.zip
  wget -q https://github.com/dbry/WavPack/releases/download/5.7.0/wavpack-5.7.0.tar.xz
  tar -xf curl-8.11.1.tar.xz && tar -xf wavpack-5.7.0.tar.xz && unzip -qo sqlite-amalgamation-3470200.zip

  mkdir -p "${DEPS}/lib" "${DEPS}/include"
  ${CC:-cc} -O2 -fPIC -DSQLITE_THREADSAFE=1 -c sqlite-amalgamation-3470200/sqlite3.c -o sqlite3.o
  ar rcs "${DEPS}/lib/libsqlite3.a" sqlite3.o
  cp sqlite-amalgamation-3470200/sqlite3.h "${DEPS}/include/"

  cmake -S wavpack-5.7.0 -B wavpack-build -G Ninja -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="${DEPS}" -DCMAKE_INSTALL_LIBDIR=lib -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    -DBUILD_SHARED_LIBS=OFF -DBUILD_TESTING=OFF -DWAVPACK_BUILD_PROGRAMS=OFF -DWAVPACK_BUILD_DOCS=OFF \
    -DWAVPACK_INSTALL_DOCS=OFF -DWAVPACK_ENABLE_LIBCRYPTO=OFF
  cmake --build wavpack-build && cmake --install wavpack-build

  cd curl-8.11.1
  ./configure --prefix="${DEPS}" --disable-shared --enable-static --with-pic --with-openssl --with-zlib \
    --without-libpsl --without-brotli --without-zstd --without-nghttp2 --without-libidn2 \
    --without-librtmp --without-libssh2 --disable-ldap --disable-ldaps --disable-manual --disable-docs \
    --without-ca-bundle --without-ca-path --with-ca-fallback
  make -j"$(nproc)" -C lib && make -C lib install && make -C include install
fi

cd "${SRC}"
# The client starts the local server as <binary dir>/DDNet-Server; ship it as DDNet-Server.<arch>
sed -i "s/#define SERVER_EXEC \"DDNet-Server[^\"]*\"/#define SERVER_EXEC \"DDNet-Server.${ARCH}\"/" src/engine/client/updater.h
grep -q "SERVER_EXEC \"DDNet-Server.${ARCH}\"" src/engine/client/updater.h || { echo "Error: SERVER_EXEC not patched"; exit 1; }
# No desktop notifications: libnotify (glib, gdk-pixbuf) is not on CFWs
sed -i '/#include <libnotify\/notify.h>/d; /#define NOTIFICATIONS_USE_LIBNOTIFY/d' src/engine/client/notifications.cpp
if grep -q libnotify src/engine/client/notifications.cpp; then echo "Error: libnotify not removed"; exit 1; fi

# GLES only: CONF_BACKEND_OPENGL_ES drops the desktop OpenGL backend (as on Android), so no GLEW
# (GLX) is needed and the client links libGLESv2 where upstream links libGL.
cmake -S . -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_CXX_FLAGS="${CXXFLAGS} -DCONF_BACKEND_OPENGL_ES=1" \
  -DCMAKE_DISABLE_FIND_PACKAGE_GLEW=ON \
  -DOPENGL_gl_LIBRARY="${LIBDIR}/libGLESv2.so" \
  -DOPENSSL_USE_STATIC_LIBS=TRUE \
  -DCURL_LIBRARY="${DEPS}/lib/libcurl.a;${LIBDIR}/libssl.a;${LIBDIR}/libcrypto.a;z;pthread;dl" \
  -DCURL_INCLUDEDIR="${DEPS}/include" \
  -DSQLite3_LIBRARY="${DEPS}/lib/libsqlite3.a" \
  -DSQLite3_INCLUDEDIR="${DEPS}/include" \
  -DWAVPACK_LIBRARY="${DEPS}/lib/libwavpack.a" \
  -DWAVPACK_INCLUDEDIR="${DEPS}/include/wavpack" \
  -DPREFER_BUNDLED_LIBS=OFF \
  -DVULKAN=OFF -DVIDEORECORDER=OFF -DINFORM_UPDATE=OFF -DAUTOUPDATE=OFF -DTOOLS=OFF \
  -DDISCORD=OFF -DUPNP=OFF -DDOWNLOAD_GTEST=OFF \
  "${CMAKE_ARGS[@]}"

${PORT_BUILD}

rm -rf dist && mkdir -p dist/licenses
cp "build/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp build/DDNet-Server "dist/DDNet-Server.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" "dist/DDNet-Server.${ARCH}"
cp -r build/data dist/
cp license.txt dist/licenses/LICENSE.ddnet.txt
cp "${DEPS}/src/curl-8.11.1/COPYING" dist/licenses/LICENSE.curl.txt
cp "${DEPS}/src/wavpack-5.7.0/COPYING" dist/licenses/LICENSE.wavpack.txt
cp /usr/share/doc/libssl-dev/copyright dist/licenses/LICENSE.openssl.txt

ldd "dist/${PORT_EXE}.${ARCH}" "dist/DDNet-Server.${ARCH}" || true
objdump -T "dist/${PORT_EXE}.${ARCH}" | grep -o 'GLIBC_[0-9.]*' | sort -uV | tail -1

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
ls -lha dist
