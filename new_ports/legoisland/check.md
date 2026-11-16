# check_port: legoisland

`check_port.sh legoisland --fix --fix-license` on 2026-11-16 17:30, commit 01cfba5.

**Result: ok, 2 warning(s)**

## Warnings

- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## Fixed by --fix

- set attr.min_glibc from "2.30" to "2.29" (libs.aarch64/liblego1.so needs GLIBC_2.29)

## Notes

- aarch64 loads from libs.aarch64: libSDL3.so.0 liblego1.so libtheoradec.so.1
- x86_64 loads from libs.x86_64: libSDL3.so.0 liblego1.so
