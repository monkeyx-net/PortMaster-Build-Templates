# check_port: soh

`check_port.sh soh --fix --fix-bloat --fix-license` on 2026-11-12 09:32, commit f477fb1.

**Result: ok, 8 warning(s)**

## Warnings

- Ship of Harkinian.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libs.x86_64/libtinyxml2.so.10 is not loaded: not NEEDED by or named in the binary or another bundled lib (drop it from FILES for x86_64 in build.sh, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libtinyxml2.so.10 is not loaded: not NEEDED by or named in the binary or another bundled lib (drop it from FILES for aarch64 in build.sh, or add it to dlopen_libs if it is dlopen'd)
- build.sh FILES copies libcrypto.so.1.1 for x86_64 but libs.x86_64/ has none and only aarch64 loads it: move it to a [[ ${ARCH} == "aarch64" ]] && FILES+=( ) block
- build.sh FILES copies libspdlog.so.1.12 for aarch64 x86_64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libfmt.so.9 for x86_64 but nothing NEEDs it: drop it from FILES

## Fixed by --fix-bloat

- remove libs.x86_64/libz.so.1: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libz.so.1: CFWs already ship it (also drop it from FILES in build.sh)

## Notes

- aarch64 loads from libs.aarch64: libSDL2_net-2.0.so.0 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libtheoradec.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libSDL2_net-2.0.so.0 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libzip.so.5
