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
  # Upstream defaults to -mcpu=native on aarch64, which tunes for the CI runner's server CPU and
  # can emit instructions Cortex-A53/A55 handhelds don't have; target baseline ARMv8 instead.
  CPU_OPTION="-march=armv8-a"
elif [[ ${ARCH} == "x86_64" ]]; then
  SOURCE_DIR="/usr/lib/x86_64-linux-gnu"
fi

#### Lighthouse ####
git clone https://github.com/HarbourMasters/Lighthouse.git
cd Lighthouse

git fetch --tags
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

git submodule update --init --recursive

# Verify where clang is or use the default 'clang' which is v18 on 24.04
if ! command -v clang-18 &> /dev/null; then
    export CC=clang
    export CXX=clang++
else
    export CC=clang-18
    export CXX=clang++-18
fi

mkdir build-lighthouse && cd build-lighthouse
cmake .. -GNinja -DUSE_OPENGLES=1 -DCMAKE_BUILD_TYPE=Release ${CPU_OPTION:+-DCPU_OPTION="${CPU_OPTION}"}
# Builds the standalone torch and packs lighthouse.o2r (bk.o2r needs the rom, so it's made on device)
cmake --build . --target GeneratePortO2R -j$(nproc)
cmake --build . -j$(nproc)
cd ../..
ls -lha Lighthouse/build-lighthouse/
if [[ ! -f "Lighthouse/build-lighthouse/${PORT_EXE}" ]]; then
  echo "Error: ${PORT_EXE} was not built"
  exit 1
fi
mkdir -p "${CDIR}/dist/assets"
strip "Lighthouse/build-lighthouse/${PORT_EXE}" || true
cp "Lighthouse/build-lighthouse/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
cp "Lighthouse/build-lighthouse/lighthouse.o2r" "dist/"
cp "Lighthouse/build-lighthouse/gamecontrollerdb.txt" "dist/" || true
# Standalone torch, for converting the rom without a window if the launcher needs it
TORCH="Lighthouse/build-lighthouse/TorchExternal/src/TorchExternal-build/torch"
strip "${TORCH}" || true
cp "${TORCH}" "dist/assets/torch.${ARCH}" || true
# The in-game extractor reads config.yml and assets/yaml/ from the app folder; ship them as
# extractor.zip to unpack into the port folder. This also replaces the LFS pointer for
# extractor.zip that the package job checks out with lfs: false.
(cd "${CDIR}/Lighthouse" && zip -qr "${CDIR}/dist/assets/extractor.zip" config.yml assets)
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
