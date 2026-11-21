# check_port: opengolf

`check_port.sh opengolf --test` on 2026-11-21 12:24, commit c00c4d4f4.

**Result: dry run, changes: 5 for --fix; 1 error(s) left, 1 warning(s)**

## Errors

- port.json genre "casual" is not one of: action, adventure, arcade, casino/card, fps, platformer, puzzle, racing, rhythm, rpg, simulation, sports, strategy, visual novel, other

## Warnings

- libcairo.so.2 (aarch64) is needed by libs.aarch64/libtheoradec.so.1 but not bundled or in cfw_provided: check ldd on devices

## To fix with --fix

- Open Golf.sh:22: change "> file" to ": > file" (SC2188)
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- shrink screenshot.png from 269 to 78 kB (128 colours)
- shrink cover.png from 269 to 78 kB (128 colours)

## Notes

- aarch64 loads from libs.aarch64: libtheoradec.so.1
