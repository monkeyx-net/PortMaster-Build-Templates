#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"
RECIPE="${PWD}/recipes/ports/${PORT_FOLDER}"
PORT_VERSION=$(jq -r '.source.port_version' "${RECIPE}/recipe.json")

# Aurora builds its disc reader (nod) from Rust source via Corrosion, and the builder images do
# not ship Rust, so install a toolchain into $HOME (HOME=/tmp inside the build container).
export CARGO_HOME="${HOME}/.cargo"
export RUSTUP_HOME="${HOME}/.rustup"
if ! command -v cargo >/dev/null 2>&1; then
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
    | sh -s -- -y --profile minimal --default-toolchain stable --no-modify-path
fi
export PATH="${CARGO_HOME}/bin:${PATH}"
cargo --version

cd "new_ports/${PORT_FOLDER}/source"

# The release tarball has empty submodule folders (extern/aurora, extern/borealis).
if [[ ! -f extern/aurora/CMakeLists.txt || ! -f extern/borealis/CMakeLists.txt ]]; then
  rm -rf /tmp/nookery-git
  git clone --depth 1 --branch "${PORT_VERSION}" --recurse-submodules --shallow-submodules \
    https://github.com/flyngmt/Nookery.git /tmp/nookery-git
  cp -a /tmp/nookery-git/extern/aurora/. extern/aurora/
  cp -a /tmp/nookery-git/extern/borealis/. extern/borealis/
  rm -rf /tmp/nookery-git
fi

# Aurora: give Dawn's EGL the window's X11 display instead of EGL_DEFAULT_DISPLAY (crusty/Mali need it)
# (-N and the grep allow re-running on an already patched source)
patch -p1 -N -r - < "${RECIPE}/aurora-gles-x11-display.patch" || true
grep -q AURORA_GLES_X11_DISPLAY extern/aurora/lib/webgpu/gpu.cpp || { echo "Error: Aurora patch not applied"; exit 1; }
# Aurora: read GX vertex/array data from r32uint textures instead of storage buffers, which libmali
# does not allow in the vertex stage (GL_MAX_VERTEX_SHADER_STORAGE_BLOCKS = 0)
patch -p1 -N -r - < "${RECIPE}/aurora-gles-vertex-textures.patch" || true
grep -q BufferTextureWidth extern/aurora/lib/gfx/resources.hpp || { echo "Error: Aurora texture patch not applied"; exit 1; }

# Dawn is built from source with only its OpenGL ES backend: PortMaster devices have no Vulkan, and
# the prebuilt Dawn packages are built on Ubuntu 24.04 (too new a glibc/libstdc++ for the target).
# SDL3 and nod are vendored and linked statically. There is no HTTP backend (update checks only).
# The source is not a git checkout, so pass the version instead of letting borealis run git describe.
# Dawn and Aurora need C++20 library features (std::bit_cast, std::atomic::wait, std::jthread) that
# the Ubuntu 20.04 libstdc++ 10 lacks, so use LLVM's libc++ (jthread is still "experimental" in
# libc++ 18) and link it statically, so the binary needs only glibc. Upstream builds the decompiled
# C with GCC; clang 16+ makes some of its leniencies errors, so keep them as warnings.
cmake -S . -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_C_COMPILER=clang-18 \
  -DCMAKE_CXX_COMPILER=clang++-18 \
  -DCMAKE_C_FLAGS="-Wno-error=return-type -Wno-error=int-conversion -Wno-error=incompatible-function-pointer-types -Wno-error=implicit-function-declaration -Wno-error=implicit-int" \
  -DCMAKE_CXX_FLAGS="-stdlib=libc++ -fexperimental-library" \
  -DCMAKE_EXE_LINKER_FLAGS="-fuse-ld=lld -nostdlib++ -Wl,-Bstatic -lc++ -lc++experimental -lc++abi -Wl,-Bdynamic" \
  -DBOREALIS_APP_VERSION_OVERRIDE="${PORT_VERSION}" \
  -DBOREALIS_HTTP_BACKEND=none \
  -DAURORA_DAWN_PROVIDER=vendor \
  -DAURORA_DAWN_LINKAGE=static \
  -DAURORA_SDL3_PROVIDER=vendor \
  -DAURORA_SDL3_LINKAGE=static \
  -DSDL_X11_XTEST=OFF \
  -DAURORA_NOD_PROVIDER=vendor \
  -DDAWN_ENABLE_VULKAN=OFF \
  -DDAWN_ENABLE_DESKTOP_GL=OFF \
  -DDAWN_ENABLE_OPENGLES=ON

# Dawn (fetched during configure) dlopens the unversioned "libEGL.so", a dev symlink most CFWs do not
# ship; under Westonpack it should also pick the same libEGL.so.1 as crusty. Load the soname instead.
DAWN_GL=build/_deps/dawn-src/src/dawn/native/opengl/BackendGL.cpp
perl -pi -e 's/const char\* kEGLLib = "libEGL" DAWN_ANGLE_LIBS_SUFFIX ".so";/const char* kEGLLib = "libEGL.so.1";/' "${DAWN_GL}"
if ! grep -q 'kEGLLib = "libEGL.so.1"' "${DAWN_GL}"; then
  echo "Error: failed to patch the EGL library name in ${DAWN_GL}"
  exit 1
fi
# EGL 1.4 drivers (e.g. Mali blobs) can list EGL_EXT_platform_base without exporting
# eglCreatePlatformPixmapSurfaceEXT, which Dawn rejects the adapter for but never calls.
DAWN_EGL=build/_deps/dawn-src/src/dawn/native/opengl/EGLFunctions.cpp
perl -pi -e 's/GET_PROC_WITH_NAME\(CreatePlatformPixmapSurface, "eglCreatePlatformPixmapSurfaceEXT"\);/CreatePlatformPixmapSurface = reinterpret_cast<decltype(CreatePlatformPixmapSurface)>(GetProcAddress("eglCreatePlatformPixmapSurfaceEXT"));/' "${DAWN_EGL}"
if grep -q 'GET_PROC_WITH_NAME(CreatePlatformPixmapSurface' "${DAWN_EGL}"; then
  echo "Error: failed to patch the optional pixmap surface proc in ${DAWN_EGL}"
  exit 1
fi
# crusty: eglChooseConfig returns the current context's config (fails once Dawn's context is current),
# so cache the window config found first; and fence syncs need the current context's EGLDisplay.
patch -d build/_deps/dawn-src -p1 -N -r - < "${RECIPE}/dawn-crusty-egl.patch" || true
grep -q mSyncDisplay build/_deps/dawn-src/src/dawn/native/opengl/UtilsEGL.cpp || { echo "Error: Dawn crusty EGL patch not applied"; exit 1; }

${PORT_BUILD}

# The game loads its UI resources (fonts, RmlUi documents, localization) from res/ next to the binary.
mkdir -p dist
cp "build/bin/${PORT_EXE}" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true
cp -r build/bin/res dist/

ldd "dist/${PORT_EXE}.${ARCH}" || true

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
ls -lha dist
