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
  "liblzma.so.5"
  "libogg.so.0"
  "libvorbis.so.0"
  "libvorbisfile.so.3"
  "libvorbisenc.so.2"
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

#### Starship ####
git clone https://github.com/HarbourMasters/Starship.git
cd Starship

git fetch --tags
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

git submodule update --init --recursive

# torch builds the whole O2R in RAM (and never frees it), which runs 1GB devices out of memory
# during rom extraction; replace its writer with one that streams entries straight to the file.
cp "${CDIR}/recipes/ports/starship/torch/ZWrapper.h" "${CDIR}/recipes/ports/starship/torch/ZWrapper.cpp" tools/Torch/src/archive/

# The standalone torch (used on device to extract the rom without a window) is built as an
# ExternalProject with no build type, i.e. unoptimised; build it as Release.
sed -i 's|CMAKE_ARGS -DCMAKE_INSTALL_PREFIX=${CMAKE_BINARY_DIR}/Torch|CMAKE_ARGS -DCMAKE_INSTALL_PREFIX=${CMAKE_BINARY_DIR}/Torch -DCMAKE_BUILD_TYPE=Release|' CMakeLists.txt

# v2.0.0 passes -mfpu=neon on aarch64, which is a 32-bit ARM option that aarch64 compilers
# reject (NEON is always on there anyway). Fixed upstream after v2.0.0.
sed -i 's|set(CPU_OPTION -mfpu=neon)|set(CPU_OPTION)|' CMakeLists.txt

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-starship && cd build-starship
# dladdr (libultraship's crash handler) lives in libdl on glibc < 2.34 (the 20.04 aarch64 builder),
# and Starship doesn't link it; append it after the objects so it isn't dropped
cmake .. -GNinja -DUSE_OPENGLES=1 -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_STANDARD_LIBRARIES="-ldl"
# Builds the standalone torch and packs starship.o2r (sf64.o2r needs the rom, so it's made on device)
cmake --build . --target GeneratePortO2R -j$(nproc)
cmake --build . -j$(nproc)
cd ../..
ls -lha Starship/build-starship/
if [[ ! -f "Starship/build-starship/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "Starship/build-starship/${PORT_EXE}" || true
cp "Starship/build-starship/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "Starship/build-starship/starship.o2r" "dist/"
TORCH="Starship/build-starship/TorchExternal/src/TorchExternal-build/torch"
strip "${TORCH}" || true
cp "${TORCH}" "dist/assets/torch.${ARCH}"
# torch reads config.yml and the asset yamls (assets/yaml/...) from the working directory;
# ship them as extractor.zip to unpack into the port folder. This also replaces the LFS pointer
# for extractor.zip that the package job checks out with lfs: false.
(cd "${CDIR}/Starship" && zip -qr "${CDIR}/dist/assets/extractor.zip" config.yml assets/yaml)
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
