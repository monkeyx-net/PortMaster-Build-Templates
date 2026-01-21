#!/bin/bash

git clone --recursive https://github.com/HarbourMasters/Shipwright.git
cd Shipwright
mkdir -p build-cmake
cd build-cmake
CC=clang-18 CXX=clang++-18 cmake .. -GNinja -DUSE_OPENGLES=1 -DBUILD_CROWD_CONTROL=0 -DCMAKE_BUILD_TYPE=Release
cmake --build . --config Release --target GenerateSohOtr
cmake --build . --config Release -j$(nproc)

cd build-cmake/soh
strip soh.elf

# - Copy `soh.elf` and `soh.otr` to `roms/ports/soh`.
# - Copy the `build-cmake/assets` folder to `ports/soh` and copy `build-cmake/ZAPD/ZAPD.out` to `ports/soh/assets/`.
# - If the build is a new version open `ports/soh/assets/otrgen` with a text editor and edit `--portVer` around Line 50 to match the release version you pulled.

