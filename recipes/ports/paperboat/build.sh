#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# Paperboat needs libzip and tinyxml2; the rest are libzip's dependencies (ldd)
FILES=(
  "liblzma.so.5"
  "libtinyxml2.so.10"
  "libzip.so.5"
)
[[ ${ARCH} == "aarch64" ]] && FILES+=(
  "libbz2.so.1"
  "libcrypto.so.1.1"
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
  CPU_OPTION="-DCPU_OPTION=-mcpu=cortex-a53"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### PaperBoat ####
git clone --branch "${TAG}" --recursive https://github.com/HarbourMasters/PaperBoat.git || exit 1
cd PaperBoat

# torch builds the whole O2R in RAM (and never frees it), which runs 1GB devices out of memory
# during rom extraction; replace its writer with one that streams entries straight to the file.
cp "${RECIPE_DIR}/torch/ZWrapper.h" "${RECIPE_DIR}/torch/ZWrapper.cpp" external/torch/src/archive/

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

# The default CPU_OPTION on aarch64 is -mcpu=native (the build machine); target the oldest
# supported SoC instead. libultraship's crash handler needs dladdr (libdl on glibc < 2.34).
cmake -S . -B build-paperboat -GNinja \
    -DUSE_OPENGLES=1 \
    ${CPU_OPTION} \
    -DCMAKE_BUILD_TYPE=${BUILD_TYPE:-Release} \
    -DCMAKE_CXX_FLAGS=-ldl
cmake --build build-paperboat --target GeneratePortO2R -j$(nproc)
cmake --build build-paperboat -j$(nproc)
cd ..
ls -lha PaperBoat/build-paperboat/
if [[ ! -f "PaperBoat/build-paperboat/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist"
strip "PaperBoat/build-paperboat/${PORT_EXE}" || true
cp "PaperBoat/build-paperboat/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "PaperBoat/build-paperboat/paperboat.o2r" "dist/"
# The game's built-in torch reads config.yml and assets/ next to the binary to convert the rom
cp PaperBoat/config.yml dist/
cp -r PaperBoat/assets dist/
ls -lha dist/

mkdir -p ${CDIR}/dist/libs.${ARCH}
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
