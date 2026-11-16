## Information
LEGO Island, portable ([isle-portable](https://github.com/isledecomp/isle-portable)) is a platform-independent version of LEGO Island (v1.1, English) built from the [isledecomp](https://github.com/isledecomp/isle) decompilation.

## Installation
You need your own copy of the LEGO Island CD (v1.1, English).

1. Copy the contents of the CD's `LEGO` folder into `ports/legoisland/LEGO`.
2. Copy the `DATA` folder from the CD's `DATA/DISK/LEGO` folder into `ports/legoisland/LEGO` as well.

`ports/legoisland/LEGO` then contains `SCRIPTS` and `DATA` (about 450 MB). Saves go to `ports/legoisland/saves`. Logs are written to `ports/legoisland/log.txt`; please include it when reporting an issue.

## Launcher
A settings launcher opens before the game: renderer, island detail and textures, frame rate limit, transitions, cursor speed, music and the third-person camera extension. Choose Play, or wait for it to start the game. Try OpenGL ES 2 or Software if the picture is wrong.

## Controls

| Button | Action |
|--|--|
| D-pad | Move |
| Left stick | Cursor |
| Right stick | Cursor |
| A | Click |
| B | Skip (Space) |
| R2 | Click |
| Select | Menu (Esc) |
| Start | Pause |

## Thanks
- The isledecomp team for the decompilation and isle-portable
- Mindscape and LEGO Media for the game
- Port by monkeyx

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [isledecomp/isle-portable](https://github.com/isledecomp/isle-portable). The recipe folder [recipes/ports/legoisland](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/legoisland) holds `build.sh`, which builds the source tarball from `recipe.json` (aarch64 in an Ubuntu 20.04 container for glibc compatibility). SDL3 and iniparser are fetched by CMake; SDL3, `liblego1.so` and the game are bundled. The OpenGL 1 and SDL GPU renderers, the Qt config tool and multiplayer are left out:

```sh
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_DISABLE_FIND_PACKAGE_OpenGL=ON \
  -DISLE_RENDERER_SDLGPU=OFF -DISLE_BUILD_CONFIG=OFF -DISLE_USE_LWS=OFF
cmake --build build
```
