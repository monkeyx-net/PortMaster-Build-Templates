# check_port: ghostship

`check_port.sh ghostship --fix` on 2026-11-18 22:08, commit 660b015f9.

**Result: ok, 3 warning(s)**

## Warnings

- Ghostship.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Fixed by --fix

- shrink screenshot.png from 535 to 93 kB (128 colours)
- shrink cover.png from 493 to 80 kB (256 colours)

## Notes

- shellcheck not installed, launcher lint skipped
- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libtcc.so libtheoradec.so.1 libtinyxml2.so.10 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libtcc.so libtinyxml2.so.10 libzip.so.5
