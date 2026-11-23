# check_port: frozenbubble

`check_port.sh frozenbubble --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 9 for --fix, 2 for --fix-license; 0 error(s) left, 4 warning(s)**

## Warnings

- no license for bundled libs.aarch64/fallback/libpcre.so.3: add licenses/LICENSE.pcre.txt
- no license for bundled libs.aarch64/fallback/libpcre2-8.so.0 libs.x86_64/fallback/libpcre2-8.so.0: add licenses/LICENSE.pcre2-8.txt
- libharfbuzz.so.0 (x86_64) is needed by libs.x86_64/libSDL2_ttf-2.0.so.0 but not bundled or in cfw_provided: check ldd on devices
- build.sh FILES copies libtheoradec.so.1 but no libs.\<arch>/ has it although aarch64 x86_64 needs it: rebuild (or it is missing on the builder)

## To fix with --fix

- set Frozen Bubble.sh to 644 (staged in git)
- Frozen Bubble.sh:22: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.28" (libs.aarch64/fallback/libglib-2.0.so.0 needs GLIBC_2.28)
- remove love_11.5 from attr.runtime (it ships with PortMaster)
- rename licenses/LICENSE_frozen-bubble.txt to LICENSE.frozen-bubble.txt
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Setup screen (Love2D launcher), ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- shrink screenshot.png from 433 to 93 kB (128 colours)
- shrink cover.png from 399 to 92 kB (128 colours)

## To fix with --fix-license

- fetch licenses/LICENSE.SDL2_ttf.txt for libs.aarch64/libSDL2_ttf-2.0.so.0 libs.x86_64/libSDL2_ttf-2.0.so.0 from https://raw.githubusercontent.com/libsdl-org/SDL_ttf/SDL2/LICENSE.txt
- fetch licenses/LICENSE.glib.txt for libs.aarch64/fallback/libglib-2.0.so.0 libs.x86_64/fallback/libglib-2.0.so.0 from https://raw.githubusercontent.com/GNOME/glib/main/COPYING

## Notes

- aarch64 loads from libs.aarch64: libglib-2.0.so.0 libpcre2-8.so.0 libpcre.so.3 libSDL2_ttf-2.0.so.0
- x86_64 loads from libs.x86_64: libglib-2.0.so.0 libpcre2-8.so.0 libSDL2_ttf-2.0.so.0
