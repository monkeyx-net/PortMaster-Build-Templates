# check_port: nookery

`check_port.sh nookery --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 4 for --fix; 0 error(s) left, 4 warning(s)**

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libzstd.so.1 (x86_64) is needed by Nookery.x86_64 but not bundled or in cfw_provided: check ldd on devices
- screenshot.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)
- cover.png is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)

## To fix with --fix

- Nookery.sh:23: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (Nookery.aarch64 needs GLIBC_2.30)
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Launcher, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
