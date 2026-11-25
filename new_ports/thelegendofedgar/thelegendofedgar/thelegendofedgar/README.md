# The Legend of Edgar

## Notes

A 2D platform game with a persistent world. When Edgar's father fails to return home after venturing out one dark and stormy night, Edgar fears the worst: he has been captured by the evil sorcerer who lives in a fortress beyond the forbidden swamp.

Thanks to [Parallel Realities](https://github.com/riksweeney/edgar) for creating this game.

## Controls

| Button | Action |
|--|--|
| Dpad | Move |
| Left Analog | Move|
| Right Analog | Move |
| A | Jump |
| B | Inventory |
| X | Use Selected Item |
| Y | Interact/Pull Item |
| L1 | Attack with equipped weapon |
| L2 | Previous Item |
| R1 | Block with equipped shield |
| R2 | Next Item |

## Compile

`make DATA_DIR=./ -j$(nproc)`


Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [riksweeney/edgar](https://github.com/riksweeney/edgar). The recipe folder [recipes/ports/thelegendofedgar](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/thelegendofedgar) holds `build.sh`, which reads `recipe.json` and applies the patches and files in that folder before running the build command:

```sh
make DATA_DIR=./ -j$(nproc)
```
