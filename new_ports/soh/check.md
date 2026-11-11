# check_port: soh

`check_port.sh soh --test` on 2026-11-11 21:50, commit d31bd1f0d.

**Result: dry run, changes: 5 for --fix, 8 for --fix-bloat, 9 for --fix-license; 2 error(s) left, 8 warning(s)**

## Errors

- gameinfo.xml root element is \<gamelist>, it must be \<gameList>
- gameinfo.xml \<releasedate> is '20220322', use YYYYMMDDT000000

## Warnings

- Ship of Harkinian.sh sets a trap: make sure pm_finish still runs
- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libopusfile.so.0 (aarch64) is needed by soh.elf.aarch64 but not bundled or in cfw_provided: check ldd on devices
- libSDL2_net-2.0.so.0 (aarch64) is needed by soh.elf.aarch64 but not bundled or in cfw_provided: check ldd on devices
- libSDL2_net-2.0.so.0 (x86_64) is needed by soh.elf.x86_64 but not bundled or in cfw_provided: check ldd on devices
- libopusfile.so.0 (x86_64) is needed by soh.elf.x86_64 but not bundled or in cfw_provided: check ldd on devices

## To fix with --fix

- set Ship of Harkinian.sh to 644 (staged in git)
- Ship of Harkinian.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (soh.elf.aarch64 needs GLIBC_2.29)
- shrink screenshot.png from 425 to 91 kB (128 colours)
- shrink cover.png from 172 to 50 kB (256 colours)

## To fix with --fix-bloat

- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libbz2.so.1: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libz.so.1: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.aarch64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libbz2.so.1: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libz.so.1: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libbz2.so.1.0: CFWs already ship it (also drop it from FILES in build.sh)

## To fix with --fix-license

- fetch licenses/LICENSE.libzip.txt for libs.aarch64/libzip.so.5 libs.x86_64/libzip.so.5 from https://raw.githubusercontent.com/nih-at/libzip/main/LICENSE
- fetch licenses/LICENSE.xz.txt for libs.aarch64/liblzma.so.5 libs.x86_64/liblzma.so.5 from https://raw.githubusercontent.com/tukaani-project/xz/master/COPYING
- fetch licenses/LICENSE.fmt.txt for libs.x86_64/libfmt.so.9 from https://raw.githubusercontent.com/fmtlib/fmt/master/LICENSE
- fetch licenses/LICENSE.tinyxml2.txt for libs.aarch64/libtinyxml2.so.10 libs.x86_64/libtinyxml2.so.10 from https://raw.githubusercontent.com/leethomason/tinyxml2/master/LICENSE.txt
- fetch licenses/LICENSE.libogg.txt for libs.aarch64/libogg.so.0 libs.x86_64/libogg.so.0 from https://raw.githubusercontent.com/xiph/ogg/master/COPYING
- fetch licenses/LICENSE.opus.txt for libs.aarch64/libopus.so.0 libs.x86_64/libopus.so.0 from https://raw.githubusercontent.com/xiph/opus/main/COPYING
- fetch licenses/LICENSE.spdlog.txt for libs.aarch64/libspdlog.so.1 libs.x86_64/libspdlog.so.1.12 from https://raw.githubusercontent.com/gabime/spdlog/v1.x/LICENSE
- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING
- fetch licenses/LICENSE.openssl.txt for libs.aarch64/libcrypto.so.1.1 from https://raw.githubusercontent.com/openssl/openssl/OpenSSL_1_1_1-stable/LICENSE
