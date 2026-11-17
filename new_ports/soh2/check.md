# check_port: soh2

`check_port.sh soh2 --fix` on 2026-11-17 19:07, commit 7b9715bb3.

**Result: ok, 7 warning(s)**

## Warnings

- Ship of Harkinian 2.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libspdlog.so.1 (aarch64) is needed by 2s2h.elf.aarch64 but not bundled or in cfw_provided: check ldd on devices
- libspdlog.so.1.12 (x86_64) is needed by 2s2h.elf.x86_64 but not bundled or in cfw_provided: check ldd on devices
- libbz2.so.1 (x86_64) is needed by 2s2h.elf.x86_64 but not bundled or in cfw_provided: check ldd on devices
- libfmt.so.9 (x86_64) is needed by 2s2h.elf.x86_64 but not bundled or in cfw_provided: check ldd on devices
- build.sh FILES copies libtheoradec.so.1 but no libs.\<arch>/ has it although aarch64 needs it: rebuild (or it is missing on the builder)

## Fixed by --fix

- set attr.min_glibc from "" to "2.29" (2s2h.elf.aarch64 needs GLIBC_2.29)
- add ## Compile section to README.md
- shrink screenshot.png from 433 to 82 kB (64 colours)
- shrink cover.png from 196 to 52 kB (256 colours)

## Notes

- shellcheck not installed, launcher lint skipped
- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libtinyxml2.so.10 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libtinyxml2.so.10 libzip.so.5
