# check_port: stuntcarracer

`check_port.sh stuntcarracer --test` on 2026-11-22 18:36, commit 5d6b25891.

**Result: dry run, changes: 3 for --fix; 0 error(s) left, 4 warning(s)**

## Warnings

- no license for bundled libs.aarch64/libGLEW.so.2.1: add licenses/LICENSE.GLEW.txt
- no license for bundled libs.aarch64/libSDL_ttf-2.0.so.0: add licenses/LICENSE.SDL_ttf.txt
- libs.aarch64/libGLEW.so.2.1 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libSDL_ttf-2.0.so.0 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)

## To fix with --fix

- Stunt Car Racer.sh:21: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (stuntcarracer.aarch64 needs GLIBC_2.29)
- shrink screenshot.png from 143 to 42 kB (256 colours)
