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

cd "new_ports/build/${PORT_FOLDER}"
git clone https://github.com/nih-at/libzip.git
cd libzip
git checkout 0581df510597b46c28509e9d4b5998cf5fecb636
mkdir build-soh && cd build-soh
cmake ..
make -j8
make install
cd ../..

git clone https://github.com/nlohmann/json.git
cd json
git checkout f3dc4684b40a124cabc8554967c2cd8db54f15dd
mkdir build-soh && cd build-soh
cmake ..
make -j8
make install
cd ../..

git clone https://github.com/libarchive/bzip2.git
cd bzip2
git checkout 1ea1ac188ad4b9cb662e3f8314673c63df95a589
mkdir build-soh && cd build-soh
cmake ..
make -j8
make install
cd ../..

git clone https://github.com/leethomason/tinyxml2.git
cd tinyxml2
git checkout 57ec94127bda7979282315b7d4b0059eeb6f3b5d
git checkout .
mkdir build-soh && cd build-soh
cmake -DBUILD_SHARED_LIBS=ON ..
make -j8
make install

# prevent this file being found by cmake when SoH is compiled
cd ..
mv cmake/tinyxml2-config.cmake cmake/tinyxml2-config.cmake.disabled
cd ..

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
cp "${PORT_EXE}" "dist/${PORT_EXE}.elf.${ARCH}"
strip "dist/${PORT_EXE}.elf.${ARCH}" || true
cp "${PORT_EXE}.o2r" "dist/"
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
