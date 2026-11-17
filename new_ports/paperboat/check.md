# check_port: paperboat

`check_port.sh paperboat --fix --fix-license` on 2026-11-17 12:22, commit cfc81d6.

**Result: ok, 5 warning(s)**

## Warnings

- PaperBoat.sh sets a trap: make sure pm_finish still runs
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## Fixed by --fix

- PaperBoat.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (Paperboat.aarch64 needs GLIBC_2.30)
- add ## Compile section to README.md

## Notes

- aarch64 loads from libs.aarch64: libbz2.so.1 libcrypto.so.1.1 liblzma.so.5 libtinyxml2.so.10 libz.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: libbz2.so.1.0 liblzma.so.5 libtinyxml2.so.10 libzip.so.5
