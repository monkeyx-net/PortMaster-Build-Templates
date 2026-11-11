# check_port: lix

`check_port.sh lix --fix --fix-bloat --fix-license` on 2026-11-11 20:42, commit 378385009.

**Result: ok, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
