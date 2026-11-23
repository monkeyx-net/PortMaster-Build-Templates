# check_port: animalcrossing

`check_port.sh animalcrossing --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 6 for --fix; 1 error(s) left, 0 warning(s)**

## Errors

- port.json genre "todo" is not one of: action, adventure, arcade, casino/card, fps, platformer, puzzle, racing, rhythm, rpg, simulation, sports, strategy, visual novel, other

## To fix with --fix

- set Unofficial Linux PC Port of Animal Crossing Gamecube.sh to 644 (staged in git)
- Unofficial Linux PC Port of Animal Crossing Gamecube.sh:22: change "> file" to ": > file" (SC2188)
- rename licenses/LICENSE_AnimalCrossing.txt to LICENSE.AnimalCrossing.txt
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- letterbox screenshot.png from 1280x576 to 640x480
