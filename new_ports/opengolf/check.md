# check_port: opengolf

`check_port.sh opengolf --fix --fix-license` on 2026-11-28 18:48, commit f9f889c.

**Result: ok, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix

- shrink screenshot.png from 269 to 84 kB (64 colours)

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
