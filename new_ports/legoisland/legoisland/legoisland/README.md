## Information
LEGO Island, portable ([isle-portable](https://github.com/isledecomp/isle-portable)) is a platform-independent version of LEGO Island (v1.1, English) built from the [isledecomp](https://github.com/isledecomp/isle) decompilation.

## Installation
You need your own copy of the LEGO Island CD (v1.1, English).

Copy the whole CD into `ports/legoisland/LEGO`, so that it contains the CD's `LEGO` and `DATA` folders. A smaller copy works too: put the CD's `LEGO/SCRIPTS` folder and its `DATA/DISK/LEGO/DATA` folder into `ports/legoisland/LEGO` (about 450 MB). Saves go to `ports/legoisland/saves`. Logs are written to `ports/legoisland/log.txt`; please include it when reporting an issue.

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

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [isledecomp/isle-portable](https://github.com/isledecomp/isle-portable). The recipe folder [recipes/ports/legoisland](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/legoisland) holds `build.sh`, which builds the source tarball from `recipe.json` (aarch64 in an Ubuntu 20.04 container for glibc compatibility). iniparser is fetched by CMake. SDL3 is [bmdhacks/SDL `sdl2-backend`](https://github.com/bmdhacks/SDL/tree/sdl2-backend) at `6057d79`, an SDL3 that runs on the CFW's own SDL2, with `sdl3-sdl2-backend-fixes.patch` from the recipe folder (absolute mouse motion, no segfault on exit). SDL3, `liblego1.so` and the game are bundled. The OpenGL 1 and SDL GPU renderers, the Qt config tool and multiplayer are left out:

```sh
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_DISABLE_FIND_PACKAGE_OpenGL=ON \
  -DISLE_RENDERER_SDLGPU=OFF -DISLE_BUILD_CONFIG=OFF -DISLE_USE_LWS=OFF \
  -DFETCHCONTENT_SOURCE_DIR_SDL3=<shim> -DSDL_SDL2_BACKEND=ON -DSDL_GPU=OFF -DSDL_RENDER_GPU=OFF \
  -DSDL_X11=OFF -DSDL_WAYLAND=OFF -DSDL_KMSDRM=OFF -DSDL_VULKAN=OFF -DSDL_PIPEWIRE=OFF -DSDL_PULSEAUDIO=OFF \
  -DSDL_ALSA=OFF -DSDL_SNDIO=OFF -DSDL_OSS=OFF -DSDL_JACK=OFF -DSDL_UNIX_CONSOLE_BUILD=ON
cmake --build build
```
