# check_port: opengolf

`check_port.sh opengolf --fix --fix-license` on 2026-11-21 12:41, commit fa61d5a.

**Result: ok, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix

- Open Golf.sh:22: change "> file" to ": > file" (SC2188)
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling, ## Compile -> ## Notes, ## Controls, ## Compile
- shrink screenshot.png from 269 to 81 kB (64 colours)
- shrink cover.png from 269 to 81 kB (64 colours)

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
