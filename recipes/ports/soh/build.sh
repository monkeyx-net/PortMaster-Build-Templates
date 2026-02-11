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

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-soh && cd build-soh
cmake .. -GNinja -DUSE_OPENGLES=1 -DBUILD_CROWD_CONTROL=0 -DCMAKE_BUILD_TYPE=Release
cmake --build . -j8
cmake --build . --target GenerateSohOtr -j8
cd ../..
mkdir -p ${CDIR}/dist/assets/extractor
ls -lha Shipwright/build-soh/soh/
strip "Shipwright/build-soh/soh/${PORT_EXE}.elf" || true
cp "Shipwright/build-soh/soh/${PORT_EXE}.elf" "dist/${PORT_EXE}.elf.${ARCH}"
cp "Shipwright/build-soh/soh/${PORT_EXE}.o2r" "dist/"
cp "Shipwright/build-soh/ZAPD/ZAPD.out" "${CDIR}/dist/assets/extractor/ZAPD.out"
ls -lha dist/
cd dist/assets
mkdir assets_zip
cp -r "${CDIR}/Shipwright/soh/assets/extractor/." assets_zip/
cp -r "${CDIR}/Shipwright/soh/assets/xml/" assets_zip/
cd assets_zip
zip -r ../extractor.zip ./*
cd ..
rm -rf assets_zip

#get license file

mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
