#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# needs file.arch?
FILES=(
  "libz.so.1"
  "libbz2.so.1"
  "libbz2.so.1.0"
  "libopus.so.0"
  "liblzma.so.5"
  "libogg.so.0"
  "libpng16.so.16"
  "libspdlog.so.1"
  "libspdlog.so.1.12"
  "libzip.so.5"
  "libtinyxml2.so.10"
  "libfmt.so.9"
  "libcrypto.so.1.1"
)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### SoH ####
git clone https://github.com/HarbourMasters/Shipwright.git
cd Shipwright

git fetch --tags
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

git submodule update --init

# Ubuntu 20.04 builder ships libstdc++-10 which lacks std::bit_cast (C++20, libstdc++-11+).
# Provide a shim backed by clang's __builtin_bit_cast.
sed -i 's|#include <bit>|#include <bit>\n#ifndef __cpp_lib_bit_cast\nnamespace std { template <class To, class From> constexpr To bit_cast(const From\& f) noexcept { return __builtin_bit_cast(To, f); } }\n#endif|' soh/soh/GbiWrap.cpp

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-soh && cd build-soh
cmake .. -GNinja -DUSE_OPENGLES=1 -DBUILD_CROWD_CONTROL=1 -DCMAKE_BUILD_TYPE=Release
cmake --build . -j$(nproc)
cmake --build . --target GenerateSohOtr -j$(nproc)
cd ../..
ls -lha Shipwright/build-soh/soh/
if [[ ! -f "Shipwright/build-soh/soh/${PORT_EXE}.elf" ]]; then
  echo "Error: ${PORT_EXE}.elf was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "Shipwright/build-soh/soh/${PORT_EXE}.elf" || true
cp "Shipwright/build-soh/soh/${PORT_EXE}.elf" "dist/${PORT_EXE}.elf.${ARCH}"
cp "Shipwright/build-soh/soh/${PORT_EXE}.o2r" "dist/"
# 9.3+: ZAPD/xml extractor replaced by torch built into soh.elf; it reads the asset ymls from ./assets
# Ship them as extractor.zip (unpacked into assets/ by otrgen). This also replaces the LFS pointer
# for extractor.zip that the package job checks out with lfs: false.
(cd "${CDIR}/Shipwright/soh/assets/yml" && zip -qr "${CDIR}/dist/assets/extractor.zip" .)
unzip -tq "${CDIR}/dist/assets/extractor.zip" || exit 1
ls -lha dist/ dist/assets/

#get license file

mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
