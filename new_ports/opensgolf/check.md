# check_port: opensgolf

`check_port.sh opensgolf --fix` on 2026-11-27 15:20, commit 4044d0180.

**Result: 1 error(s), 3 warning(s)**

## Errors

- gameinfo.xml \<image> ./opensgolf/cover.png does not exist

## Warnings

- no aarch64 binaries in new_ports/opensgolf/opensgolf/opensgolf (not built yet?)
- no screenshot.png (make_images.sh creates a placeholder)
- no cover.png (make_images.sh creates a placeholder)

## Fixed by --fix

- add ## Compile section to README.md
