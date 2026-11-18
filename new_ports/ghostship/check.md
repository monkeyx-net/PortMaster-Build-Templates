# check_port: ghostship

`check_port.sh ghostship --test` on 2026-11-18 16:30, commit b64a98a86.

**Result: dry run, changes: 1 for --fix; 1 error(s) left, 4 warning(s)**

## Errors

- gameinfo.xml \<image> ./ghostship/cover.png does not exist

## Warnings

- Ghostship.sh sets a trap: make sure pm_finish still runs
- no aarch64 binaries in new_ports/ghostship/ghostship/ghostship (not built yet?)
- no screenshot.png (make_images.sh creates a placeholder)
- no cover.png (make_images.sh creates a placeholder)

## To fix with --fix

- add ## Compile section to README.md

## Notes

- shellcheck not installed, launcher lint skipped
