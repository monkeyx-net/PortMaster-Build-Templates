# check_port: lix

`check_port.sh lix --test` on 2026-11-11 20:31, commit e13946fef.

**Result: dry run, changes: 4 for --fix, 6 for --fix-bloat, 1 for --fix-license; 1 error(s) left, 2 warning(s)**

## Errors

- gameinfo.xml is missing \<releasedate>

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libbrotlidec.so.1 (x86_64) is needed by libs.x86_64/libfreetype.so.6 but not bundled or in cfw_provided: check ldd on devices

## To fix with --fix

- Lix.sh:22: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (libs.aarch64/libpng16.so.16 needs GLIBC_2.29)
- shrink screenshot.png from 218 to 55 kB (256 colours)
- shrink cover.png from 190 to 64 kB (256 colours)

## To fix with --fix-bloat

- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libfreetype.so.6: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)

## To fix with --fix-license

- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING
