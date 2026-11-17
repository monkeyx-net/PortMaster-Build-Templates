# check_port: lighthouse

`check_port.sh lighthouse --fix --fix-license` on 2026-11-17 16:58, commit 55ac608.

**Result: ok, 8 warning(s)**

## Warnings

- Lighthouse.sh sets a trap: make sure pm_finish still runs
- bundles libs.x86_64/libpng16.so.16: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- bundles libs.aarch64/libpng16.so.16: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libs.x86_64/libpng16.so.16 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libpng16.so.16 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)
- libs.aarch64/libspdlog.so.1 is not loaded: not NEEDED by or named in the binary or another bundled lib, and not in FILES in build.sh (stale from an older build, or copied some other way: remove it, or add it to dlopen_libs if it is dlopen'd)

## Fixed by --fix

- set Lighthouse.sh to 644 (staged in git)
- Lighthouse.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (Lighthouse.aarch64 needs GLIBC_2.30)
- add ## Compile section to README.md
- shrink screenshot.png from 329 to 97 kB (128 colours)
- shrink cover.png from 331 to 95 kB (64 colours)

## Fixed by --fix-license

- fetch licenses/LICENSE.libzip.txt for libs.x86_64/libzip.so.5 libs.aarch64/libzip.so.5 from https://raw.githubusercontent.com/nih-at/libzip/main/LICENSE
- fetch licenses/LICENSE.xz.txt for libs.x86_64/liblzma.so.5 libs.aarch64/liblzma.so.5 from https://raw.githubusercontent.com/tukaani-project/xz/master/COPYING
- fetch licenses/LICENSE.libogg.txt for libs.x86_64/libogg.so.0 libs.aarch64/libogg.so.0 from https://raw.githubusercontent.com/xiph/ogg/master/COPYING
- fetch licenses/LICENSE.zlib.txt for libs.x86_64/libz.so.1 libs.aarch64/libz.so.1 from https://raw.githubusercontent.com/madler/zlib/master/LICENSE
- fetch licenses/LICENSE.spdlog.txt for libs.aarch64/libspdlog.so.1 from https://raw.githubusercontent.com/gabime/spdlog/v1.x/LICENSE
- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING
- fetch licenses/LICENSE.bzip2.txt for libs.x86_64/libbz2.so.1 libs.x86_64/libbz2.so.1.0 libs.aarch64/libbz2.so.1 libs.aarch64/libbz2.so.1.0 from https://gitlab.com/bzip2/bzip2/-/raw/master/COPYING
- fetch licenses/LICENSE.SDL2_net.txt for libs.x86_64/libSDL2_net-2.0.so.0 libs.aarch64/libSDL2_net-2.0.so.0 from https://raw.githubusercontent.com/libsdl-org/SDL_net/SDL2/LICENSE.txt
- fetch licenses/LICENSE.libpng.txt for libs.x86_64/libpng16.so.16 libs.aarch64/libpng16.so.16 from https://raw.githubusercontent.com/pnggroup/libpng/libpng16/LICENSE
- fetch licenses/LICENSE.openssl.txt for libs.aarch64/libcrypto.so.1.1 from https://raw.githubusercontent.com/openssl/openssl/OpenSSL_1_1_1-stable/LICENSE
- fetch licenses/LICENSE.libvorbis.txt for libs.x86_64/libvorbisfile.so.3 libs.x86_64/libvorbisenc.so.2 libs.x86_64/libvorbis.so.0 libs.aarch64/libvorbisfile.so.3 libs.aarch64/libvorbisenc.so.2 libs.aarch64/libvorbis.so.0 from https://raw.githubusercontent.com/xiph/vorbis/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libSDL2_net-2.0.so.0 libbz2.so.1 libbz2.so.1.0 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libtheoradec.so.1 libtinyxml2.so.10 libvorbis.so.0 libvorbisenc.so.2 libvorbisfile.so.3 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libSDL2_net-2.0.so.0 libbz2.so.1 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libtinyxml2.so.10 libvorbis.so.0 libvorbisenc.so.2 libvorbisfile.so.3 libz.so.1 libzip.so.5
