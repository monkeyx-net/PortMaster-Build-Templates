# check_port: soh2

`check_port.sh soh2 --test` on 2026-11-12 12:18, commit 79329d6ec.

**Result: dry run, changes: 7 for --fix, 2 for --fix-bloat, 10 for --fix-license; 1 error(s) left, 3 warning(s)**

## Errors

- gameinfo.xml is not valid XML: no element found: line 14, column 10

## Warnings

- shellcheck new_ports/soh2/soh2/Ship of Harkinian 2.sh:88:20: warning: Declare and assign separately to avoid masking return values. [SC2155]
- shellcheck new_ports/soh2/soh2/Ship of Harkinian 2.sh:93:28: warning: Quote this to prevent word splitting. [SC2046]
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## To fix with --fix

- Ship of Harkinian 2.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.29" (2s2h.elf.aarch64 needs GLIBC_2.29)
- rename licenses/LICENSE_OTRExporter.txt to LICENSE.OTRExporter.txt
- rename licenses/LICENSE_LibUltraShip.txt to LICENSE.LibUltraShip.txt
- rename licenses/LICENSE_ZAPDTR.txt to LICENSE.ZAPDTR.txt
- shrink screenshot.png from 433 to 82 kB (64 colours)
- shrink cover.png from 196 to 52 kB (256 colours)

## To fix with --fix-bloat

- remove libs.aarch64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)
- remove libs.x86_64/libpng16.so.16: CFWs already ship it (also drop it from FILES in build.sh)

## To fix with --fix-license

- fetch licenses/LICENSE.opus.txt for libs.aarch64/libopusfile.so.0 libs.x86_64/libopusfile.so.0 from https://raw.githubusercontent.com/xiph/opus/main/COPYING
- fetch licenses/LICENSE.libzip.txt for libs.aarch64/libzip.so.5 libs.x86_64/libzip.so.5 from https://raw.githubusercontent.com/nih-at/libzip/main/LICENSE
- fetch licenses/LICENSE.xz.txt for libs.aarch64/liblzma.so.5 libs.x86_64/liblzma.so.5 from https://raw.githubusercontent.com/tukaani-project/xz/master/COPYING
- fetch licenses/LICENSE.fmt.txt for libs.x86_64/libfmt.so.9 from https://raw.githubusercontent.com/fmtlib/fmt/master/LICENSE
- fetch licenses/LICENSE.tinyxml2.txt for libs.aarch64/libtinyxml2.so.10 libs.x86_64/libtinyxml2.so.10 from https://raw.githubusercontent.com/leethomason/tinyxml2/master/LICENSE.txt
- fetch licenses/LICENSE.libogg.txt for libs.aarch64/libogg.so.0 libs.x86_64/libogg.so.0 from https://raw.githubusercontent.com/xiph/ogg/master/COPYING
- fetch licenses/LICENSE.zlib.txt for libs.aarch64/libz.so.1 libs.x86_64/libz.so.1 from https://raw.githubusercontent.com/madler/zlib/master/LICENSE
- fetch licenses/LICENSE.opus.txt for libs.aarch64/libopus.so.0 libs.x86_64/libopus.so.0 from https://raw.githubusercontent.com/xiph/opus/main/COPYING
- fetch licenses/LICENSE.spdlog.txt for libs.aarch64/libspdlog.so.1 libs.x86_64/libspdlog.so.1.12 from https://raw.githubusercontent.com/gabime/spdlog/v1.x/LICENSE
- fetch licenses/LICENSE.openssl.txt for libs.aarch64/libcrypto.so.1.1 from https://raw.githubusercontent.com/openssl/openssl/OpenSSL_1_1_1-stable/LICENSE

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libpng16.so.16 libspdlog.so.1 libtinyxml2.so.10 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1 libbz2.so.1.0 libfmt.so.9 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libpng16.so.16 libspdlog.so.1.12 libtinyxml2.so.10 libzip.so.5 libz.so.1
