# check_port: manicminer

`check_port.sh manicminer --test` on 2026-11-22 18:35, commit 5d6b25891.

**Result: dry run, changes: 2 for --fix; 3 error(s) left, 0 warning(s)**

## Errors

- manicminer.x86_64 is x86_64 but port.json attr.arch does not list x86_64
- manicminer.armhf is armhf but port.json attr.arch does not list armhf
- no licenses/ folder

## To fix with --fix

- set Manic Miner.sh to 644 (staged in git)
- Manic Miner.sh:22: change "> file" to ": > file" (SC2188)
