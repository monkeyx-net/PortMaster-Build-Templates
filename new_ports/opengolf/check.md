# check_port: opengolf

`check_port.sh opengolf --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: none; 0 error(s) left, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
