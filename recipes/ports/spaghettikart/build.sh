#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# Spaghettify needs ogg/vorbis, libzip and tinyxml2; the rest are libzip's dependencies (ldd)
FILES=(
  "liblzma.so.5"
  "libogg.so.0"
  "libvorbis.so.0"
  "libvorbisenc.so.2"
  "libvorbisfile.so.3"
  "libtinyxml2.so.10"
  "libzip.so.5"
)
[[ ${ARCH} == "aarch64" ]] && FILES+=(
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

if [[ ${ARCH} ==  "aarch64" ]]; then
  SOURCE_DIR="/usr/lib/aarch64-linux-gnu"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### SpaghettiKart ####
# The recipe pins a main commit (port_download = archive/<sha>.tar.gz); clone it because the
# tarball has no submodules
SHA=$(jq -r '.source.port_download' "${RECIPE_DIR}/recipe.json" | sed -n 's|.*/archive/\([0-9a-f]\{40\}\)\.tar\.gz$|\1|p')
if [[ -z "${SHA}" ]]; then
  echo "Error: no commit in the recipe's port_download"
  exit 1
fi
git clone https://github.com/HarbourMasters/SpaghettiKart.git
cd SpaghettiKart
git checkout "${SHA}" || exit 1
git submodule update --init --recursive

# libultraship's SegAddr treats any odd address as segmented and indexes the segment table with
# its top bits, so a native 64-bit pointer with bit 0 set (e.g. a resource path string) crashes
# with SIGSEGV on arm64. Pass anything above the N64 segment range through unchanged.
# LUS PR #1004 did this but was closed without being merged.
sed -i '/^void\* Interpreter::SegAddr(uintptr_t w1) {$/a\    if (w1 > 0x0FFFFFFF) {\n        return (void*)w1;\n    }' libultraship/src/fast/interpreter.cpp
sed -i 's|uint32_t segNum = (uint32_t)(w1 >> 24);|uintptr_t segNum = w1 >> 24;|; s|uint32_t offset = w1 & 0x00FFFFFE;|uintptr_t offset = w1 \& 0x00FFFFFE;|' libultraship/src/fast/interpreter.cpp
if ! grep -q 'if (w1 > 0x0FFFFFFF) {' libultraship/src/fast/interpreter.cpp \
   || ! grep -q 'uintptr_t segNum = w1 >> 24;' libultraship/src/fast/interpreter.cpp; then
  echo "Error: SegAddr patch did not apply"
  exit 1
fi

# torch builds the whole O2R in RAM (and never frees it), which runs 1GB devices out of memory
# during rom extraction; replace its writer with one that streams entries straight to the file.
cp "${RECIPE_DIR}/torch/ZWrapper.h" "${RECIPE_DIR}/torch/ZWrapper.cpp" torch/src/archive/

# The standalone torch (used on device to extract the rom without a window) is built as an
# ExternalProject with no build type, i.e. unoptimised; build it as Release.
sed -i 's|CMAKE_ARGS -DCMAKE_INSTALL_PREFIX=${CMAKE_BINARY_DIR}/torch|CMAKE_ARGS -DCMAKE_INSTALL_PREFIX=${CMAKE_BINARY_DIR}/torch -DCMAKE_BUILD_TYPE=Release|' CMakeLists.txt
grep -q 'torch -DCMAKE_BUILD_TYPE=Release' CMakeLists.txt || { echo "Error: torch Release patch did not apply"; exit 1; }

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-spaghettikart && cd build-spaghettikart
# dladdr (libultraship's crash handler) lives in libdl on glibc < 2.34 (the 20.04 aarch64 builder);
# append it after the objects so it isn't dropped
cmake .. -GNinja -DUSE_OPENGLES=1 -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD_LIBRARIES="-ldl"
# Builds the standalone torch and packs spaghetti.o2r (mk64.o2r needs the rom, so it's made on device)
cmake --build . --target GenerateO2R -j$(nproc)
cmake --build . -j$(nproc)
cd ../..
ls -lha SpaghettiKart/build-spaghettikart/
if [[ ! -f "SpaghettiKart/build-spaghettikart/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "SpaghettiKart/build-spaghettikart/${PORT_EXE}" || true
cp "SpaghettiKart/build-spaghettikart/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "SpaghettiKart/build-spaghettikart/spaghetti.o2r" "dist/"
TORCH="SpaghettiKart/build-spaghettikart/TorchExternal/src/TorchExternal-build/torch"
strip "${TORCH}" || true
cp "${TORCH}" "dist/assets/torch.${ARCH}"
# torch reads config.yml, the asset yamls, include/ (enums) and meta/mods.toml from the working
# directory; ship them as extractor.zip to unpack into the port folder
(cd "${CDIR}/SpaghettiKart" && zip -qr "${CDIR}/dist/assets/extractor.zip" config.yml yamls include meta)
unzip -tq "${CDIR}/dist/assets/extractor.zip" || exit 1
ls -lha dist/ dist/assets/

mkdir -p ${CDIR}/dist/libs.${ARCH}
# if sourcedir !null and files !null
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
