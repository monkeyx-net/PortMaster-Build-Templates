# check_port: soh2

`check_port.sh soh2 --fix --fix-license` on 2026-11-20 23:05, commit c00c4d4.

**Result: ok, 2 warning(s)**

## Warnings

- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libtheoradec.so.1 libtinyxml2.so.10 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libbz2.so.1 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopus.so.0 libopusfile.so.0 libtinyxml2.so.10 libzip.so.5
