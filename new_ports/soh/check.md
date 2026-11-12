# check_port: soh

`check_port.sh soh --fix --fix-bloat --fix-license` on 2026-11-12 11:56, commit 0e8c7b0.

**Result: ok, 5 warning(s)**

## Warnings

- Ship of Harkinian.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libs.x86_64/libtinyxml2.so.10 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libtinyxml2.so.10 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)

## Notes

- aarch64 loads from libs.aarch64: libSDL2_net-2.0.so.0 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libtheoradec.so.1 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libSDL2_net-2.0.so.0 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libzip.so.5
