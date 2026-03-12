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
  "libopusfile.so.0"
  "libogg.so.0"
  "libpng16.so.16"
  "libspdlog.so.1"
  "libspdlog.so.1.12"
  "libzip.so.5"
  "libtinyxml2.so.10"
  "libcrypto.so.1.1"
  "libfmt.so.9"
  "liblzma.so.5"
)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### SoH ####
git clone https://github.com/HarbourMasters/2ship2harkinian.git
cd 2ship2harkinian

git fetch --tags
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

git submodule update --init

cd libultraship
patch -p1 < $CDIR/recipes/ports/soh2/soh2.patch
cd ..

export CFLAGS="-O3 -Wall -Wextra -Wno-return-type -Wno-unused-parameter -Wno-unused-function -Wno-unused-variable -Wno-macro-redefined -Wno-unknown-warning-option"
export CXXFLAGS="-O3 -Wall -Wextra -Wno-return-type -Wno-unused-parameter -Wno-unused-function -Wno-unused-variable -Wno-macro-redefined -Wno-unknown-warning-option"

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-soh2 && cd build-soh2
cmake .. -GNinja -DUSE_OPENGLES=1 -DBUILD_CROWD_CONTROL=0 -DCMAKE_BUILD_TYPE=Release
cmake --build . -j8
cmake --build . --target Generate2ShipOtr -j8
cd ../..
mkdir -p ${CDIR}/dist/assets/extractor
ls -lha 2ship2harkinian/build-soh2/mm/
strip "2ship2harkinian/build-soh2/mm/${PORT_EXE}.elf" || true
cp "2ship2harkinian/build-soh2/mm/${PORT_EXE}.elf" "dist/${PORT_EXE}.elf.${ARCH}"
cp "2ship2harkinian/build-soh2/mm/2ship.o2r" "dist/"
cp -r "${CDIR}/2ship2harkinian/mm/assets/extractor/." "${CDIR}/dist/assets/extractor/"
cp -r "${CDIR}/2ship2harkinian/mm/assets/xml/." "${CDIR}/dist/assets/xml/"
ls -lha dist/
cd "${CDIR}/dist/assets"
zip -r ${CDIR}/dist/assets/extractor.zip ./*
rm -rf extractor/* xml
cp "${CDIR}/2ship2harkinian/build-soh2/ZAPD/ZAPD.out" "${CDIR}/dist/assets/extractor/ZAPD.out.${ARCH}"


mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
