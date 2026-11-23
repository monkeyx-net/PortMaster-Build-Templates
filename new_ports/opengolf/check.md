# check_port: opengolf

`check_port.sh opengolf --fix --fix-license` on 2026-11-23 19:47, commit 2468dbc.

**Result: ok, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix

- shrink screenshot.png from 269 to 83 kB (64 colours)

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
