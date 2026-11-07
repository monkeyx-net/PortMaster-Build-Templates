## Notes
Thanks to [FlyingMeta](https://github.com/flyngmt) for Nookery, the PC port of Animal Crossing, and to the [ACreTeam](https://github.com/ACreTeam/ac-decomp) for the decompilation it is built on.

[https://github.com/flyngmt/Nookery](https://github.com/flyngmt/Nookery)

You need your own copy of the game: place an Animal Crossing (USA, GAFE01 Rev 0) disc image (`.rvz`, `.iso`, `.gcm` or `.ciso`) in `ports/nookery`. Saves and settings go to `ports/nookery/conf`.

Nookery renders through Dawn (WebGPU). This port uses Dawn's OpenGL ES backend (GLES 3.1 required) and runs under the Westonpack runtime.

## Controls

| Button | Action |
|--|--|
| Left Analog / Dpad | Move (Control Stick) / D-Pad |
| Right Analog | C-Stick |
| A (bottom) | A |
| B (right) | B |
| X (left) | X |
| Y (top) | Y |
| R1 | Z |
| L2 / R2 | L / R |
| Start | Start |
| Select + Start | Quit |

Buttons can be remapped in the game's settings menu.

## Compiling

GitHub Actions automatic build: `recipes/ports/nookery` (CMake + Ninja, clang-18; Dawn built from source with only the OpenGL ES backend, SDL3 and nod vendored and linked statically).
