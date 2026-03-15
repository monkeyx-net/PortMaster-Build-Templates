#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# needs file.arch?
FILES=(
  "libz.so.1"
  "libzip.so.5"
  "libogg.so.0"
  "libvorbis.so.0"
  "libvorbisfile.so.3"
  "libvorbisenc.so.2"
  "libcrypto.so.1.1"

)
CDIR=$(pwd)
DEST_DIR="${CDIR}/dist/libs.${ARCH}"

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi


#### SpaghettiKart ####
git clone https://github.com/HarbourMasters/SpaghettiKart.git
cd SpaghettiKart

# Checkout latest tag
git fetch --tags
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

# Grab SpaghettiKart PR Need and IF statement here as only for aarch64?
git fetch origin pull/650/head:pr-650-spaghettikart
git checkout pr-650-spaghettikart

# Initialize and sync submodules
git submodule sync
git submodule update --init --recursive

# libultraship PR
cd libultraship
git remote add kenix https://github.com/Kenix3/libultraship.git 2>/dev/null || true
git fetch kenix pull/1004/head:pr-1004-libultraship
git checkout pr-1004-libultraship
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

## Build
mkdir build-spaghettikart && cd build-spaghettikart
cmake .. -GNinja \
  -DUSE_OPENGLES=1 \
  -DCMAKE_BUILD_TYPE=Release 

cmake --build . -j$(nproc) --config Release --target GenerateO2R
cmake --build . -j$(nproc)
cd ../..


mkdir -p ${CDIR}/dist/tools
ls -lha SpaghettiKart/build-spaghettikart/
strip "SpaghettiKart/build-spaghettikart/${PORT_EXE}" || true
cp "SpaghettiKart/build-spaghettikart/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "SpaghettiKart/build-spaghettikart/${PORT_EXE}.pdb" "dist/"
cp "SpaghettiKart/build-spaghettikart/spaghetti.o2r" "dist/"
wget -O "dist/gamecontrollerdb.txt" "https://raw.githubusercontent.com/mdqinc/SDL_GameControllerDB/master/gamecontrollerdb.txt"
cp -r "${CDIR}/SpaghettiKart/build-spaghettikart/TorchExternal/src/TorchExternal-build/torch" "${CDIR}/dist/tools/torch.${ARCH}"

rm -rf "${CDIR}/tools/yamls"
rm -f "${CDIR}/tools/config.yml"
cp -r "${CDIR}/SpaghettiKart/config.yml/." "${CDIR}/dist/tools"
ls -lha dist/

# use a subshell to avoid changing the current working directory of the main script
(
  cd "${CDIR}/SpaghettiKart" || exit
  zip -r "${CDIR}/tools/assets.zip" yamls/ meta/ include/
)


mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
