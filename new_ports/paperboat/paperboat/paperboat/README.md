## Information
PaperBoat is the HarbourMasters PC port of Paper Mario, built from the decompilation, with enhancements such as higher frame rates, widescreen support and mod support.

## Installation
You need to provide your own rom. Copy your Paper Mario US `.z64` rom (SHA1 `3837f44cda784b466c9a2d99df70d77c322b97a0`) to `ports/paperboat`. On first launch the game offers to convert it to `pm64.o2r`: answer Yes to each prompt (D-pad and A). The conversion can take a few minutes, after which the rom can be deleted.

Mods can be added to the `ports/paperboat/mods` folder. Logs are written to `ports/paperboat/log.txt`; please include it when reporting an issue.

## Controls

| Button | Action |
|--|--|
| A | A |
| B | B |
| L1 | L |
| L2 | Z |
| R2 | R |
| Right stick | C buttons |
| Start | Start |
| Select | Menu |
| L3 | Menu bar |
| R3 | Toggle alternate assets |

All buttons can be remapped from the in-game menu.

## Thanks
- HarbourMasters (Bass3l, JeodC, KiritoDv, Caladius) for PaperBoat, the native PC port
- The Paper Mario decompilation team
- Nintendo and Intelligent Systems for the game
- Port by SLAYKIN6_TTV and monkeyx

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [HarbourMasters/PaperBoat](https://github.com/HarbourMasters/PaperBoat). The recipe folder [recipes/ports/paperboat](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/paperboat) holds `build.sh`, which reads `recipe.json` and applies the patches and files in that folder before running the build command:

```sh
cmake
```
