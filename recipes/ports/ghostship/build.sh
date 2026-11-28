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
  # (".eh_frame_hdr refers to overlapping FDEs"); link everything (incl. libtcc.so) with lld instead
  LINKER_OPTIONS=(-DCMAKE_EXE_LINKER_FLAGS=-fuse-ld=lld -DCMAKE_SHARED_LINKER_FLAGS=-fuse-ld=lld -DCMAKE_MODULE_LINKER_FLAGS=-fuse-ld=lld)
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

# libultraship only enables gamepad navigation while the menu is open, so the rom conversion
# prompts can't be answered with a controller. Also enable it while any popup is open, checked
# every frame (as in JeodC's libultraship fork that PaperBoat uses).
sed -i '/^void Gui::DrawMenu() {$/a\    if (mConsoleVariable->GetInteger(CVAR_IMGUI_CONTROLLER_NAV, 0) \&\&\n        (GetMenuOrMenubarVisible() || ImGui::IsPopupOpen(nullptr, ImGuiPopupFlags_AnyPopupId | ImGuiPopupFlags_AnyPopupLevel))) {\n        mImGuiIo->ConfigFlags |= ImGuiConfigFlags_NavEnableGamepad;\n    } else {\n        mImGuiIo->ConfigFlags \&= ~ImGuiConfigFlags_NavEnableGamepad;\n    }' libultraship/src/ship/window/gui/Gui.cpp
grep -q 'ImGuiPopupFlags_AnyPopupLevel' libultraship/src/ship/window/gui/Gui.cpp || { echo "Error: popup navigation patch did not apply"; exit 1; }
# ImGui's SDL2 backend only listens to the first controller SDL lists (AutoFirst), which under
# gptokeyb can be the grabbed physical pad instead of the virtual one; listen to all of them.
sed -i 's|^\( *\)ImGui_ImplSDL2_InitForOpenGL(.*);$|&\n\1ImGui_ImplSDL2_SetGamepadMode(ImGui_ImplSDL2_GamepadMode_AutoAll);|' libultraship/src/fast/Fast3dGui.cpp
grep -q 'ImGui_ImplSDL2_GamepadMode_AutoAll' libultraship/src/fast/Fast3dGui.cpp || { echo "Error: gamepad mode patch did not apply"; exit 1; }
# libultraship takes the controller added/removed events off SDL's queue before ImGui sees them,
# so ImGui's controller list is never refreshed and the menu can't be opened or used with a pad.
# Refresh it whenever libultraship handles one of those events.
H=libultraship/src/ship/controller/physicaldevice/SDLAddRemoveDeviceEventHandler.cpp
sed -i 's|^#include <SDL2/SDL.h>$|&\n#include <imgui_impl_sdl2.h>|' "$H"
sed -i 's|^\( *\)mControlDeck->GetConnectedPhysicalDeviceManager()->HandlePhysicalDevice[A-Za-z]*(event.cdevice.which);$|&\n\1ImGui_ImplSDL2_SetGamepadMode(ImGui_ImplSDL2_GamepadMode_AutoAll);|' "$H"
[[ $(grep -c 'ImGui_ImplSDL2_SetGamepadMode' "$H") == 2 ]] || { echo "Error: gamepad refresh patch did not apply"; exit 1; }

# Backport of upstream PR #291 (after 3.0.0): a negative float angle cast straight to u16 reads past
# the sine table on ARM and crashes, e.g. a Piranha Plant's bubble (bhv_piranha_plant_bubble_loop)
sed -i 's|gSineTable\[(u16) (x) >> 4\]|gSineTable[(u16) (s32) (x) >> 4]|; s|gCosineTable\[(u16) (x) >> 4\]|gCosineTable[(u16) (s32) (x) >> 4]|' src/engine/math_util.h
[[ $(grep -c '(u16) (s32) (x) >> 4' src/engine/math_util.h) == 2 ]] || { echo "Error: sins/coss patch did not apply"; exit 1; }

# On aarch64 the game is built with -mcpu=native (the build machine); target the oldest
# supported SoC instead.
sed -i 's|set(CPU_OPTION -mcpu=native)|set(CPU_OPTION -mcpu=cortex-a53)|' CMakeLists.txt
grep -q 'set(CPU_OPTION -mcpu=cortex-a53)' CMakeLists.txt || { echo "Error: CPU_OPTION patch did not apply"; exit 1; }

# The GLES shaders use mediump floats, which Panfrost (Mesa) runs as real fp16 while libmali
# doesn't; use highp as libultraship already does for Emscripten.
G=libultraship/src/fast/backends/gfx_opengl.cpp
sed -i 's|"#version 300 es\\nprecision mediump float;"|"#version 300 es\\nprecision highp float;"|' "$G"
grep -q 'precision mediump float;' "$G" && { echo "Error: shader precision patch did not apply"; exit 1; }

# The standalone torch (packs ghostship.o2r and converts the rom on device) is built as an ExternalProject
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
# CMP0126=NEW: Ghostship names libultraship's CVars gSettings.* (cmake/lus-cvars.cmake), but with
# the policy unset libultraship's cache defaults (gControlNav, gAdvancedResolution, ...) win, so
# the launcher's settings (controller navigation, aspect ratio, VSync, low-res) were ignored.
cmake -S . -B build-ghostship -GNinja \
    -DCMAKE_POLICY_DEFAULT_CMP0126=NEW \
    -DUSE_OPENGLES=1 \
    -DUSE_SATELLA=OFF \
    -DCMAKE_DISABLE_FIND_PACKAGE_Vulkan=ON \
    "${LINKER_OPTIONS[@]}" \
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
# Standalone torch: the launcher converts the rom to sm64.o2r with it before the game starts,
# so the game's own conversion prompts (which can't be answered with a controller) never show
TORCH="Ghostship/build-ghostship/TorchExternal/src/TorchExternal-build/torch"
strip "${TORCH}" || true
cp "${TORCH}" "dist/assets/torch.${ARCH}" || exit 1
# libtcc.so (linked by the game) and the TCC runtime (headers, libtcc1.a) it uses to compile
# C mods at runtime
mkdir -p "${DEST_DIR}"
cp "Ghostship/build-ghostship/libultraship/libtcc.so" "${DEST_DIR}/" || exit 1
cp -r "Ghostship/build-ghostship/.tcc" "dist/" || exit 1
# The build also stages the builder's glibc crti.o/crtn.o; TCC only links them for executables
# and DLLs, never for mods compiled into memory (TCC_OUTPUT_MEMORY, ScriptLoader.cpp)
rm -f dist/.tcc/lib/crti.o dist/.tcc/lib/crtn.o
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
