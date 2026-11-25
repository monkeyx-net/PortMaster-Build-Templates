# check_port: onsyuri

`check_port.sh onsyuri --fix --fix-license` on 2026-11-25 08:06, commit 529ddc0.

**Result: ok, 3 warning(s)**

## Warnings

- libbz2.so.1 (aarch64) is needed by onsyuri.aarch64 but not bundled or in cfw_provided: check ldd on devices
- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## Fixed by --fix

- set attr.min_glibc from "" to "2.29" (onsyuri.aarch64 needs GLIBC_2.29)
