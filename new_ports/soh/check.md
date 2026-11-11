# check_port: soh

`check_port.sh soh --test` on 2026-11-11 23:36, commit 286a38035.

**Result: dry run, changes: none; 0 error(s) left, 7 warning(s)**

## Warnings

- Ship of Harkinian.sh sets a trap: make sure pm_finish still runs
- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices
- libzstd.so.1 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- libcrypto.so.3 (x86_64) is needed by libs.x86_64/libzip.so.5 but not bundled or in cfw_provided: check ldd on devices
- build.sh FILES copies libcrypto.so.1.1 for x86_64 but libs.x86_64/ has none and only aarch64 loads it: move it to a [[ ${ARCH} == "aarch64" ]] && FILES+=( ) block
- build.sh FILES copies libopusfile.so.0 but no libs.\<arch>/ has it although aarch64 x86_64 needs it: rebuild (or it is missing on the builder)
- build.sh FILES copies libbz2.so.1.0 but no libs.\<arch>/ has it although x86_64 needs it: rebuild (or it is missing on the builder)

## Notes

- aarch64 loads from libs.aarch64: libcrypto.so.1.1 liblzma.so.5 libogg.so.0 libopus.so.0 libSDL2_net-2.0.so.0 libtheoradec.so.1 libzip.so.5
- x86_64 loads from libs.x86_64: liblzma.so.5 libogg.so.0 libopus.so.0 libSDL2_net-2.0.so.0 libzip.so.5
