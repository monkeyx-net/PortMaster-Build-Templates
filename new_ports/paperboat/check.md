# check_port: paperboat

`check_port.sh paperboat --test` on 2026-11-17 12:35, commit bf553d62a.

**Result: dry run, changes: none; 0 error(s) left, 6 warning(s)**

## Warnings

- PaperBoat.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- build.sh FILES copies libtheoradec.so.1 but no libs.\<arch>/ has it although aarch64 needs it: rebuild (or it is missing on the builder)
- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## Notes

- shellcheck not installed, launcher lint skipped
- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libtinyxml2.so.10 libzip.so.5 libz.so.1
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libtinyxml2.so.10 libzip.so.5
