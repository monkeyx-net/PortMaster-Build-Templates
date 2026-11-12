# check_port: lighthouse

`check_port.sh lighthouse --test` on 2026-11-12 12:18, commit 79329d6ec.

**Result: dry run, changes: 5 for --fix, 2 for --fix-bloat, 9 for --fix-license; 0 error(s) left, 8 warning(s)**

## Warnings

- Lighthouse.sh sets a trap: make sure pm_finish still runs
- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libs.aarch64/libpng16.so.16 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libspdlog.so.1 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.x86_64/libpng16.so.16 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- build.sh FILES copies libSDL2_net-2.0.so.0 but no libs.\<arch>/ has it although aarch64 x86_64 needs it: rebuild (or it is missing on the builder)

## To fix with --fix

- set Lighthouse.sh to 644 (staged in git)
- Lighthouse.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (Lighthouse.aarch64 needs GLIBC_2.30)
- shrink screenshot.png from 329 to 77 kB (128 colours)
- shrink cover.png from 331 to 84 kB (128 colours)

## To fix with --fix-bloat

- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)

## To fix with --fix-license

- fetch licenses/LICENSE.libzip.txt for libs.aarch64/libzip.so.5 libs.x86_64/libzip.so.5 from https://raw.githubusercontent.com/nih-at/libzip/main/LICENSE
- fetch licenses/LICENSE.xz.txt for libs.aarch64/liblzma.so.5 libs.x86_64/liblzma.so.5 from https://raw.githubusercontent.com/tukaani-project/xz/master/COPYING
- fetch licenses/LICENSE.libogg.txt for libs.aarch64/libogg.so.0 libs.x86_64/libogg.so.0 from https://raw.githubusercontent.com/xiph/ogg/master/COPYING
- fetch licenses/LICENSE.zlib.txt for libs.aarch64/libz.so.1 libs.x86_64/libz.so.1 from https://raw.githubusercontent.com/madler/zlib/master/LICENSE
- fetch licenses/LICENSE.spdlog.txt for libs.aarch64/libspdlog.so.1 from https://raw.githubusercontent.com/gabime/spdlog/v1.x/LICENSE
- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING
- fetch licenses/LICENSE.bzip2.txt for libs.aarch64/libbz2.so.1 libs.aarch64/libbz2.so.1.0 libs.x86_64/libbz2.so.1 libs.x86_64/libbz2.so.1.0 from https://gitlab.com/bzip2/bzip2/-/raw/master/COPYING
- fetch licenses/LICENSE.openssl.txt for libs.aarch64/libcrypto.so.1.1 from https://raw.githubusercontent.com/openssl/openssl/OpenSSL_1_1_1-stable/LICENSE
- fetch licenses/LICENSE.libvorbis.txt for libs.aarch64/libvorbisfile.so.3 libs.aarch64/libvorbis.so.0 libs.aarch64/libvorbisenc.so.2 libs.x86_64/libvorbisfile.so.3 libs.x86_64/libvorbis.so.0 libs.x86_64/libvorbisenc.so.2 from https://raw.githubusercontent.com/xiph/vorbis/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libtheoradec.so.1 libtinyxml2.so.10 libvorbisenc.so.2 libvorbisfile.so.3 libvorbis.so.0 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libtinyxml2.so.10 libvorbisenc.so.2 libvorbisfile.so.3 libvorbis.so.0 libzip.so.5 libz.so.1
