## Notes

### Information

2Ship2Harkinian is the HarbourMasters PC port of Majora's Mask, with enhancements such as higher frame rates, widescreen support and mod support.

### Installation

You need to provide your own rom. It must have one of these SHA1 hashes (check it at https://2ship.equipment):

```
d6133ace5afaa0882cf214cf88daba39e266c078 - N64 US
9743aa026e9269b339eb0e3044cd5830a440c1fd - GC US
```

Put the rom (`.z64`, `.n64` or `.v64`) in `ports/soh2/baseroms` and start the port. On first run it is converted to `mm.o2r`, which can take several minutes.

Texture packs and mods go in the `ports/soh2/mods` folder. Logs are written to `ports/soh2/log.txt`; please include it when reporting an issue. HarbourMasters is not affiliated with PortMaster and does not support this distribution.

### Launcher

A settings launcher opens before the game: frame rate, match display rate, VSync, N64 low-res mode and aspect ratio. Choose Play, or wait for it to start the game.

### Menu Navigation

2Ship2Harkinian has built-in controller navigation for the menu. Press `SELECT` to open it and use the `D-PAD` to choose a submenu, then press `A` to switch focus to it. Press `B` to back out of a submenu.

### Suggested Mods

You can find mods at https://gamebanana.com/games/20371.

The [MM Reloaded](https://github.com/GhostlyDark/MM-Reloaded-2S2H) texture pack works; the HD version is enough for retro handhelds. Add the pack to `soh2/mods` and toggle alternate assets with R3 (`TAB`).

### Thanks

- Nintendo for the game
- HarbourMasters for the native PC port
- Jeod for the original port
- beniamino for the build steps allowing this to run on older Ubuntu-based systems
- AkerHasReawakened for the cover art
- Testers and Devs from the PortMaster Discord

## Controls

The port uses SDL controller mapping and controls can be remapped from the menu. For devices without a right analog stick, `HOTKEY + ABXY` gives the C-Buttons.

| Button | Action |
|--|--|
| A | A |
| B | B |
| X | C-Up (Hotkey+X) |
| Y | C-Left (Hotkey+Y) |
| L1 | L |
| L2 | Z |
| R2 | R |
| Select | Menu |
| Start | Start |
| R3 | Toggle alternate assets |

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [2ship2harkinian/2ship2harkinian](https://github.com/2ship2harkinian/2ship2harkinian). The recipe folder [recipes/ports/soh2](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/soh2) holds `build.sh`, which reads `recipe.json` and applies the patches and files in that folder before running the build command:

```sh
cmake
```
