# check_port: starship

`check_port.sh starship --fix --fix-license` on 2026-11-19 17:46, commit 79d803a.

**Result: ok, 3 warning(s)**

## Warnings

- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libs.aarch64/libspdlog.so.1 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libtheoradec.so.1 libtinyxml2.so.10 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libbz2.so.1 libbz2.so.1.0 liblzma.so.5 libtinyxml2.so.10 libz.so.1 libzip.so.5
