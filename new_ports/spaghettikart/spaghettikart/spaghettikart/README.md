## Notes

### Information

SpaghettiKart is the HarbourMasters PC port of Mario Kart 64, built from the decompilation, with enhancements such as higher frame rates, widescreen support and mod support.

### Installation

You need to provide your own rom. Copy your Mario Kart 64 US `.z64` rom to `ports/spaghettikart/baseroms`. The SHA-1 must be `579C48E211AE952530FFC8738709F078D5DD215E`; other versions (and `.n64`/`.v64` byte orders) will not work. On first launch the rom is converted to `mk64.o2r`, which can take a few minutes.

Mods can be added to the `ports/spaghettikart/mods` folder. Logs are written to `ports/spaghettikart/log.txt`; please include it when reporting an issue.

### Launcher

A settings launcher opens before the game: frame rate and level of detail. Choose Play, or wait for it to start the game.

### Thanks

- HarbourMasters for SpaghettiKart, the native PC port
- The Mario Kart 64 decompilation team
- Nintendo for the game
- Port by beniamino and monkeyx

## Controls

| Button | Action |
|--|--|
| A | Accelerate |
| B | C-Down |
| X | Brake |
| Y | C-Left |
| L1 | L |
| R1 | Hop / drift |
| L2 | Use item |
| R2 | Hop / drift |
| Select | Menu |
| Start | Pause |

X is the N64 B button (brake, or reverse when stopped). Right stick up and right are C-Up and C-Right. L3 also opens the menu, R3 toggles alternate (HD) assets from mods. All buttons can be remapped from the in-game menu.

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [https://github.com/HarbourMasters/SpaghettiKart/archive/refs/heads/main.tar.gz](https://github.com/HarbourMasters/SpaghettiKart/archive/refs/heads/main.tar.gz). The recipe folder [recipes/ports/spaghettikart](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/spaghettikart) holds `build.sh`, which reads `recipe.json` and applies the patches and files in that folder before running the build command:

```sh
cmake
```
