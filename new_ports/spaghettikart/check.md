# check_port: spaghettikart

`check_port.sh spaghettikart --fix --fix-license` on 2026-11-17 15:25, commit 2e0161e.

**Result: ok, 3 warning(s)**

## Warnings

- SpaghettiKart.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix

- SpaghettiKart.sh:29: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (Spaghettify.aarch64 needs GLIBC_2.29)
- add ## Compile section to README.md

## Fixed by --fix-license

- fetch licenses/LICENSE.libzip.txt for libs.x86_64/libzip.so.5 libs.aarch64/libzip.so.5 from https://raw.githubusercontent.com/nih-at/libzip/main/LICENSE
- fetch licenses/LICENSE.xz.txt for libs.x86_64/liblzma.so.5 libs.aarch64/liblzma.so.5 from https://raw.githubusercontent.com/tukaani-project/xz/master/COPYING
- fetch licenses/LICENSE.libogg.txt for libs.x86_64/libogg.so.0 libs.aarch64/libogg.so.0 from https://raw.githubusercontent.com/xiph/ogg/master/COPYING
- fetch licenses/LICENSE.zlib.txt for libs.x86_64/libz.so.1 libs.aarch64/libz.so.1 from https://raw.githubusercontent.com/madler/zlib/master/LICENSE
- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING
- fetch licenses/LICENSE.bzip2.txt for libs.x86_64/libbz2.so.1.0 libs.aarch64/libbz2.so.1 libs.aarch64/libbz2.so.1.0 from https://gitlab.com/bzip2/bzip2/-/raw/master/COPYING
- fetch licenses/LICENSE.openssl.txt for libs.aarch64/libcrypto.so.1.1 from https://raw.githubusercontent.com/openssl/openssl/OpenSSL_1_1_1-stable/LICENSE
- fetch licenses/LICENSE.libvorbis.txt for libs.x86_64/libvorbisfile.so.3 libs.x86_64/libvorbisenc.so.2 libs.x86_64/libvorbis.so.0 libs.aarch64/libvorbisfile.so.3 libs.aarch64/libvorbisenc.so.2 libs.aarch64/libvorbis.so.0 from https://raw.githubusercontent.com/xiph/vorbis/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libtheoradec.so.1 libtinyxml2.so.10 libvorbis.so.0 libvorbisenc.so.2 libvorbisfile.so.3 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libogg.so.0 libtinyxml2.so.10 libvorbis.so.0 libvorbisenc.so.2 libvorbisfile.so.3 libz.so.1 libzip.so.5
