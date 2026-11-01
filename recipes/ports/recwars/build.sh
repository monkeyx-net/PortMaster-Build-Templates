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
${PORT_BUILD}

# Assets (data/, maps/) are embedded in the binary via include_bytes!,
# so only the executable needs to be shipped.
mkdir -p dist
cp "target/release/rec-wars" "dist/${PORT_EXE}.${ARCH}"
strip "dist/${PORT_EXE}.${ARCH}" || true

tar -czf "/workspace/${PORT_FOLDER}-linux-${ARCH}.tar.gz" -C dist .
ls -lha dist
