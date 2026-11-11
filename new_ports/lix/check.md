# check_port: lix

`check_port.sh lix --fix` on 2026-11-11 21:10, commit 24c7eb4.

**Result: ok, 9 warning(s)**

## Warnings

- bundles libs.x86_64/libpng16.so.16: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.x86_64/libbz2.so.1.0: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.x86_64/libfreetype.so.6: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.aarch64/libpng16.so.16: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.aarch64/libbz2.so.1.0: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.aarch64/libfreetype.so.6: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- no license for bundled libs.x86_64/libbz2.so.1.0 libs.aarch64/libbz2.so.1.0: add licenses/LICENSE.bzip2.txt (--fix-license fetches it)
- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libbrotlidec.so.1 (x86_64) is needed by libs.x86_64/libfreetype.so.6 but not bundled or in cfw_provided: check ldd on devices
