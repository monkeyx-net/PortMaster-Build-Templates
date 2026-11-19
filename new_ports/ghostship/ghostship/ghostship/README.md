## Notes
Ghostship is the HarbourMasters PC port of Super Mario 64, built from the decompilation, with enhancements such as higher frame rates, widescreen support and mod support.

You need to provide your own rom. Copy your Super Mario 64 US (SHA1 `9bef1128717f958171a4afac3ed78ee2bb4e86ce`) or JP (SHA1 `8a20a5c83d6ceb0f0506cfc9fa20d8f438cafe51`) `.z64` rom to `ports/ghostship/baseroms`. On first launch, after you choose Play, it is converted to `sm64.o2r` before the game starts. This can take a few minutes, after which the rom can be deleted (keep it if you want `sm64.o2r` rebuilt automatically after a port update).

A settings launcher opens before the game: frame rate, N64 low-res mode, VSync, texture filter, log level and aspect ratio. Choose Play, or wait for it to start the game.

Mods can be added to the `ports/ghostship/mods` folder. Logs are written to `ports/ghostship/log.txt`; please include it when reporting an issue.

Thanks to HarbourMasters (KiritoDv, Lywx and contributors) for Ghostship, the Super Mario 64 decompilation team, and Nintendo for the game. Port by monkeyx.

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
| R3 | Toggle alternate assets |

All buttons can be remapped from the in-game menu.

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [HarbourMasters/Ghostship](https://github.com/HarbourMasters/Ghostship). The recipe folder [recipes/ports/ghostship](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/ghostship) holds `build.sh`, which applies the patches in that folder before building:

```sh
git clone --branch 3.0.0 --recursive https://github.com/HarbourMasters/Ghostship.git
cd Ghostship
cmake -S . -B build-ghostship -GNinja -DCMAKE_POLICY_DEFAULT_CMP0126=NEW -DUSE_OPENGLES=1 \
    -DUSE_SATELLA=OFF -DCMAKE_BUILD_TYPE=Release
cmake --build build-ghostship --target GeneratePortO2R
cmake --build build-ghostship
```
