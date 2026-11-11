# check_port: soh

`check_port.sh soh --fix --fix-bloat --fix-license` on 2026-11-11 23:59, commit 167a497.

**Result: ok, 11 warning(s)**

## Warnings

- Ship of Harkinian.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libXext.so.6 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libfontconfig.so.1 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libxcb-render.so.0 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libxcb-shm.so.0 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libpixman-1.so.0 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libxcb.so.1 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- libXrender.so.1 (aarch64) is needed by libs.aarch64/libcairo.so.2 but not bundled or in cfw_provided: check ldd on devices
- build.sh FILES copies libcrypto.so.1.1 for x86_64 but libs.x86_64/ has none and only aarch64 loads it: move it to a [[ ${ARCH} == "aarch64" ]] && FILES+=( ) block

## Fixed by --fix-license

- fetch licenses/LICENSE.cairo.txt for libs.aarch64/libcairo.so.2 from https://gitlab.freedesktop.org/cairo/cairo/-/raw/master/COPYING
- fetch licenses/LICENSE.bzip2.txt for libs.x86_64/libbz2.so.1.0 from https://gitlab.com/bzip2/bzip2/-/raw/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libSDL2_net-2.0.so.0 libcairo.so.2 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libtheoradec.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libSDL2_net-2.0.so.0 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libzip.so.5
