#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
FILES=(
  "liblzma.so.5"
  "libogg.so.0"
  "libopus.so.0"
  "libopusfile.so.0"
  "libtinyxml2.so.10"
  "libzip.so.5"
)
[[ ${ARCH} == "aarch64" ]] && FILES+=(
  "libbz2.so.1"
  "libcrypto.so.1.1"
  "libtheoradec.so.1"
  "libz.so.1"
)
[[ ${ARCH} == "x86_64" ]] && FILES+=(
  "libbz2.so.1.0"
)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"
RECIPE_DIR="${CDIR}/recipes/ports/${PORT_FOLDER}"
TAG=$(jq -r '.source.port_version' "${RECIPE_DIR}/recipe.json")

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### SoH2 ####
git clone --branch "${TAG}" https://github.com/HarbourMasters/2ship2harkinian.git || exit 1
cd 2ship2harkinian
git submodule update --init

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
cmake --build . -j$(nproc)
cmake --build . --target Generate2ShipOtr -j$(nproc)
cd ../..
ls -lha 2ship2harkinian/build-soh2/mm/
if [[ ! -f "2ship2harkinian/build-soh2/mm/${PORT_EXE}.elf" ]]; then
  echo "Error: ${PORT_EXE}.elf was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "2ship2harkinian/build-soh2/mm/${PORT_EXE}.elf" || true
cp "2ship2harkinian/build-soh2/mm/${PORT_EXE}.elf" "dist/${PORT_EXE}.elf.${ARCH}"
cp "2ship2harkinian/build-soh2/mm/2ship.o2r" "dist/"
# The rom is converted in game (ZAPD is built into 2s2h.elf); it reads Config_*.xml, filelists/
# and xml/ from assets/. Ship them as extractor.zip (unpacked into assets/ by the launcher). This
# also replaces the LFS pointer for extractor.zip that the package job checks out with lfs: false.
(cd "${CDIR}/2ship2harkinian/mm/assets/extractor" && zip -qr "${CDIR}/dist/assets/extractor.zip" .)
(cd "${CDIR}/2ship2harkinian/mm/assets" && zip -qr "${CDIR}/dist/assets/extractor.zip" xml)
unzip -tq "${CDIR}/dist/assets/extractor.zip" || exit 1
ls -lha dist/ dist/assets/

mkdir -p ${CDIR}/dist/libs.${ARCH}
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
