# check_port: paperboat

`check_port.sh paperboat --test` on 2026-11-17 11:49, commit 3236e5f06.

**Result: dry run, changes: 1 for --fix; 1 error(s) left, 11 warning(s)**

## Errors

- gameinfo.xml \<image> ./paperboat/cover.png does not exist

## Warnings

- PaperBoat.sh sets a trap: make sure pm_finish still runs
- build.sh FILES copies liblzma.so.5 for aarch64 x86_64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libtinyxml2.so.10 for aarch64 x86_64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libzip.so.5 for aarch64 x86_64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libbz2.so.1 for aarch64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libcrypto.so.1.1 for aarch64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libz.so.1 for aarch64 but nothing NEEDs it: drop it from FILES
- build.sh FILES copies libbz2.so.1.0 for x86_64 but nothing NEEDs it: drop it from FILES
- no aarch64 binaries in new_ports/paperboat/paperboat/paperboat (not built yet?)
- no screenshot.png (make_images.sh creates a placeholder)
- no cover.png (make_images.sh creates a placeholder)

## To fix with --fix

- add ## Compile section to README.md

## Notes

- shellcheck not installed, launcher lint skipped
