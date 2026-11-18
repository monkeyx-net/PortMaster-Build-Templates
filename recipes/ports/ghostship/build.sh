#!/bin/bash

#set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
# Ghostship needs libzip and tinyxml2; the rest are libzip's dependencies (ldd)
FILES=(
  "liblzma.so.5"
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
  # The 20.04 builder's GNU ld (binutils 2.34) fails the final link on clang-18 objects
  # (".eh_frame_hdr refers to overlapping FDEs"); link with lld instead
  LINKER_OPTION="-DCMAKE_EXE_LINKER_FLAGS=-fuse-ld=lld"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### Ghostship ####
git clone --branch "${TAG}" --recursive https://github.com/HarbourMasters/Ghostship.git || exit 1
cd Ghostship

# torch builds the whole O2R in RAM (and never frees it), which runs 1GB devices out of memory
# during rom extraction; replace its writer with one that streams entries straight to the file.
cp "${RECIPE_DIR}/torch/ZWrapper.h" "${RECIPE_DIR}/torch/ZWrapper.cpp" Torch/src/archive/

# The game only looks for roms next to the binary; also look in baseroms/ like the other HM ports
sed -i 's|{ installPath, Ship::Context::GetAppDirectoryPath("sm64") }|{ installPath, installPath + "/baseroms", Ship::Context::GetAppDirectoryPath("sm64") }|' src/port/Engine.cpp
grep -q 'installPath + "/baseroms"' src/port/Engine.cpp || { echo "Error: baseroms patch did not apply"; exit 1; }

# On aarch64 the game is built with -mcpu=native (the build machine); target the oldest
# supported SoC instead.
sed -i 's|set(CPU_OPTION -mcpu=native)|set(CPU_OPTION -mcpu=cortex-a53)|' CMakeLists.txt
grep -q 'set(CPU_OPTION -mcpu=cortex-a53)' CMakeLists.txt || { echo "Error: CPU_OPTION patch did not apply"; exit 1; }

# The standalone torch (only used to pack ghostship.o2r here) is built as an ExternalProject
# with no build type, i.e. unoptimised; build it as Release.
sed -i 's|CMAKE_ARGS -DCMAKE_INSTALL_PREFIX=${CMAKE_BINARY_DIR}/Torch|& -DCMAKE_BUILD_TYPE=Release|' CMakeLists.txt

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

# USE_SATELLA=OFF: the online relay / mod signing needs mbedtls 3 and ixwebsocket, and
# PortMaster devices are mostly offline. Vulkan is off so the binary only needs GLES.
# libultraship's crash handler needs dladdr (libdl on glibc < 2.34).
cmake -S . -B build-ghostship -GNinja \
    -DUSE_OPENGLES=1 \
    -DUSE_SATELLA=OFF \
    -DCMAKE_DISABLE_FIND_PACKAGE_Vulkan=ON \
    ${LINKER_OPTION} \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_STANDARD_LIBRARIES="-ldl"
cmake --build build-ghostship --target GeneratePortO2R -j$(nproc)
cmake --build build-ghostship -j$(nproc)
cd ..
ls -lha Ghostship/build-ghostship/
if [[ ! -f "Ghostship/build-ghostship/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "Ghostship/build-ghostship/${PORT_EXE}" || true
cp "Ghostship/build-ghostship/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "Ghostship/build-ghostship/ghostship.o2r" "dist/"
# libtcc.so (linked by the game) and the TCC runtime (headers, libtcc1.a) it uses to compile
# C mods at runtime
mkdir -p "${DEST_DIR}"
cp "Ghostship/build-ghostship/libultraship/libtcc.so" "${DEST_DIR}/" || exit 1
cp -r "Ghostship/build-ghostship/.tcc" "dist/" || exit 1
# The game's built-in torch reads config.yml and the asset yamls (assets/ymls) next to the
# binary; ship them as extractor.zip to unpack into the port folder.
(cd "${CDIR}/Ghostship" && zip -qr "${CDIR}/dist/assets/extractor.zip" config.yml assets)
unzip -tq "${CDIR}/dist/assets/extractor.zip" || exit 1
ls -lha dist/ dist/assets/

mkdir -p ${CDIR}/dist/libs.${ARCH}
for file in "${FILES[@]}"; do
    cp "${SOURCE_DIR}/${file}" "${DEST_DIR}/" 2>/dev/null || echo "Warning: ${file} not found"
done

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C ${CDIR}/dist .
pwd
ls -lha
