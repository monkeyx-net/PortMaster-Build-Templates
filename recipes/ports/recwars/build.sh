#!/bin/bash

set -e

PORT_FOLDER="$1"
PORT_BUILD="$2"
PORT_EXE="$3"
ARCH="$4"

# The builder images do not ship Rust, so install a toolchain into $HOME
# (HOME=/tmp inside the build container).
export CARGO_HOME="${HOME}/.cargo"
export RUSTUP_HOME="${HOME}/.rustup"
if ! command -v cargo >/dev/null 2>&1; then
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
    | sh -s -- -y --profile minimal --default-toolchain stable --no-modify-path
fi
export PATH="${CARGO_HOME}/bin:${PATH}"
cargo --version

cd "new_ports/${PORT_FOLDER}/source"

# miniquad dlopens "libX11.so" (and libXi/libGL/libEGL) before the versioned soname. On CFWs that
# ship the unversioned dev symlink this loads the system libX11 alongside Westonpack's libX11.so.6
# (used by crusty/SDL), two copies share one Display and the game segfaults
# ("Xlib: request 55 length 16 would exceed buffer size"). Prefer the versioned names instead.
cargo fetch --locked
MINIQUAD_SRC=$(ls -d "${CARGO_HOME}"/registry/src/*/miniquad-*/src | head -n 1)
grep -rlZ 'Module::load("lib' "${MINIQUAD_SRC}" | xargs -0 perl -0pi -e \
  's/load\(\s*"(lib[\w-]+\.so)"\s*\)(\s*\.or_else\(\|_\|\s*[\w:]*load\(\s*)"(\1\.[\d.]+)"/load("$3")$2"$1"/g'
if grep -rq 'Module::load("lib[A-Za-z0-9_-]*\.so")\s*$' "${MINIQUAD_SRC}"; then
  echo "Error: failed to patch miniquad library names in ${MINIQUAD_SRC}"
  exit 1
fi

${PORT_BUILD}

# Assets (data/, maps/) are embedded in the binary via include_bytes!,
# so only the executable needs to be shipped.
mkdir -p dist
cp "target/release/rec-wars" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
ls -lha dist
