# check_port: spaghettikart

`check_port.sh spaghettikart --test` on 2026-11-19 17:26, commit d7c6a8bb5.

**Result: dry run, changes: none; 0 error(s) left, 2 warning(s)**

## Warnings

- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libtheoradec.so.1 libtinyxml2.so.10 libvorbisenc.so.2 libvorbisfile.so.3 libvorbis.so.0 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libogg.so.0 libtinyxml2.so.10 libvorbisenc.so.2 libvorbisfile.so.3 libvorbis.so.0 libzip.so.5 libz.so.1
