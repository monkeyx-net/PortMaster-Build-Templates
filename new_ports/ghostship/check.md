# check_port: ghostship

`check_port.sh ghostship --fix --fix-license` on 2026-11-28 16:58, commit 4f4b757.

**Result: ok, 2 warning(s)**

## Warnings

- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libtcc.so libtheoradec.so.1 libtinyxml2.so.10 libzip.so.5
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libtcc.so libtinyxml2.so.10 libzip.so.5
