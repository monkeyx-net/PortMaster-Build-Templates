# check_port: ddnet

`check_port.sh ddnet --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 5 for --fix, 1 for --fix-license; 0 error(s) left, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## To fix with --fix

- DDNet.sh:22: change "> file" to ": > file" (SC2188)
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- shrink screenshot.png from 119 to 38 kB (256 colours)
- shrink cover.png from 119 to 38 kB (256 colours)

## To fix with --fix-license

- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
