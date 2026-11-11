# check_port: lix

`check_port.sh lix --fix --fix-bloat --fix-license` on 2026-11-11 21:29, commit 3c07e03.

**Result: ok, 1 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix-bloat

- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)
