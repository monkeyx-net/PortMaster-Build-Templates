# check_port: legoisland

`check_port.sh legoisland --fix` on 2026-11-16 11:33, commit 709d191.

**Result: 1 error(s), 3 warning(s)**

## Errors

- gameinfo.xml \<image> ./legoisland/cover.png does not exist

## Warnings

- no aarch64 binaries in new_ports/legoisland/legoisland/legoisland (not built yet?)
- no screenshot.png (make_images.sh creates a placeholder)
- no cover.png (make_images.sh creates a placeholder)

## Fixed by --fix

- remove love_11.5 from attr.runtime (it ships with PortMaster)
- add ## Compile section to README.md

## Notes

- shellcheck not installed, launcher lint skipped
