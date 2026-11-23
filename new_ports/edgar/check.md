# check_port: edgar

`check_port.sh edgar --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 6 for --fix; 3 error(s) left, 0 warning(s)**

## Errors

- port.json lists "Edgar.sh" but it is not in new_ports/edgar/edgar
- "The Legend of Edgar.sh" is in new_ports/edgar/edgar but not in port.json items
- launcher Edgar.sh not found

## To fix with --fix

- set attr.min_glibc from "" to "2.29" (edgar.aarch64 needs GLIBC_2.29)
- rename licenses/LICENSE_edgar.txt to LICENSE.edgar.txt
- add ## Compile section to README.md
- README.md sections: ## Notes, ## Controls, ## Compiling -> ## Notes, ## Controls, ## Compile
- shrink screenshot.png from 102 to 30 kB (256 colours)
- shrink cover.png from 113 to 34 kB (256 colours)
