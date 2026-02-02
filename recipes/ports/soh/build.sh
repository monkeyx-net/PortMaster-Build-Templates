#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
DEST_DIR="dist/libs.${ARCH}"
FILES=(
  "libz.so.1"
  "libpng16.so.16"
  "libspdlog.so.1"
)


if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu/"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu/"
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
cmake .. -GNinja -DUSE_OPENGLES=1 \
 -DBUILD_CROWD_CONTROL=0 -DCMAKE_BUILD_TYPE=Release
cmake --build . -j8

cmake --build . --target GenerateSohOtr -j8


mkdir -p dist/libs.${ARCH}
cp "Shipwright/build-soh/soh/${PORT_EXE}.elf" "dist/${PORT_EXE}.elf.${ARCH}"
#strip "dist/${PORT_EXE}.elf.${ARCH}" || true
cp "Shipwright/build-soh/soh/${PORT_EXE}.o2r" "dist/"
mkdir -p ./assets/extractor
cp "Shipwright/build-soh/ZAPD/ZAPD.out" dist/assets/extractor/ZAPD.out


#get license file

# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/"
done


tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
pwd
ls -lha
