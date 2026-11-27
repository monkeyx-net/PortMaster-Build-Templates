# check_port: opensgolf

`check_port.sh opensgolf --fix --fix-license` on 2026-11-27 15:56, commit 886f94a.

**Result: ok, 0 warning(s)**

## Fixed by --fix

- set attr.min_glibc from "" to "2.30" (opensgolf.aarch64 needs GLIBC_2.30)
- shrink screenshot.png from 464 to 80 kB (32 colours)
- shrink cover.png from 464 to 80 kB (32 colours)
