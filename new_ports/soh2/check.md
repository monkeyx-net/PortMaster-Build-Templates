# check_port: soh2

`check_port.sh soh2 --test` on 2026-11-17 19:43, commit 9004f5d40.

**Result: dry run, changes: none; 0 error(s) left, 3 warning(s)**

## Warnings

- Ship of Harkinian 2.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices

## Notes

- shellcheck not installed, launcher lint skipped
- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libtheoradec.so.1 libtinyxml2.so.10 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1 libbz2.so.1.0 liblzma.so.5 libogg.so.0 libopusfile.so.0 libopus.so.0 libtinyxml2.so.10 libzip.so.5
