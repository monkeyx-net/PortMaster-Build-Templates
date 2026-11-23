# check_port: recwars

`check_port.sh recwars --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 9 for --fix; 0 error(s) left, 1 warning(s)**

## Warnings

- licenses/ names should be LICENSE.\<component>.txt: agpl-3.0.txt 

## To fix with --fix

- set RecWars.sh to 644 (staged in git)
- RecWars.sh:23: change "> file" to ": > file" (SC2188)
- set attr.min_glibc from "" to "2.30" (recwars.aarch64 needs GLIBC_2.30)
- rename licenses/LICENSE_recwars.txt to LICENSE.recwars.txt
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- letterbox screenshot.png from 640x354 to 640x480
- shrink screenshot.png from 344 to 64 kB (128 colours)
- shrink cover.png from 344 to 64 kB (128 colours)
