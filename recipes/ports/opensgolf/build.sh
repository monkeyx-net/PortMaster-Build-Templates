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

# miniquad (vendored in vendor/miniquad) dlopens "libX11.so" (and libXi/libGL/libEGL) before the
# versioned soname. On CFWs that ship the unversioned dev symlink this loads the system libX11
# alongside Westonpack's libX11.so.6, two copies share one Display and the game segfaults
# ("Xlib: request 55 length 16 would exceed buffer size"). Prefer the versioned names instead.
MINIQUAD_SRC=vendor/miniquad/src/native
grep -rlZ '"lib[A-Za-z0-9_-]*\.so"' "${MINIQUAD_SRC}" | xargs -0 perl -0pi -e \
  's/"(lib[\w-]+\.so)"(\s*\)?\s*(?:\.or_else\(\|_\|\s*[\w:]*load\(\s*|,\s*))"(\1\.[\d.]+)"/"$3"$2"$1"/g'
if grep -rqPz '"(lib(?:X11|Xi|GL|EGL)\.so)"\s*\)?\s*(?:\.or_else\(\|_\|\s*[\w:]*load\(\s*|,\s*)"\1\.' "${MINIQUAD_SRC}"; then
  echo "Error: failed to patch miniquad library names in ${MINIQUAD_SRC}"
  exit 1
fi

# Westonpack's crusty cannot draw the system cursor over this game, so always draw the game's own
# controller pointer (an arrow, drawn last each frame) at the mouse position (gptokeyb's mouse).
perl -0pi -e 's/(fn draw_gamepad_pointer\(&mut self\) \{)\s*if !self\.pad\.active \{\s*return;\s*\}/$1\n        let pointer = if self.pad.active { self.pad.pointer } else { self.mouse };/; s/to_virtual\(self\.pad\.pointer\.0, self\.pad\.pointer\.1\)/to_virtual(pointer.0, pointer.1)/' crates/simgolf/src/gamepad.rs
if ! grep -q 'to_virtual(pointer.0, pointer.1)' crates/simgolf/src/gamepad.rs; then
  echo "Error: failed to patch the pointer in crates/simgolf/src/gamepad.rs"
  exit 1
fi

${PORT_BUILD}

# The free base set, fonts and text are built into the binary, so only the executable is shipped.
mkdir -p dist
cp "target/release/opensgolf" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
ls -lha dist
