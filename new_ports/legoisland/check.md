# check_port: legoisland

`check_port.sh legoisland --fix --fix-license` on 2026-11-16 13:04, commit dacd1ca.

**Result: ok, 2 warning(s)**

## Warnings

- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## Fixed by --fix

- LEGO Island.sh:28: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (libs.aarch64/libSDL3.so.0 needs GLIBC_2.30)

## Fixed by --fix-license

- fetch licenses/LICENSE.libtheora.txt for libs.aarch64/libtheoradec.so.1 from https://raw.githubusercontent.com/xiph/theora/master/COPYING

## Notes

- aarch64 loads from libs.aarch64: libSDL3.so.0 liblego1.so libtheoradec.so.1
- x86_64 loads from libs.x86_64: libSDL3.so.0 liblego1.so
