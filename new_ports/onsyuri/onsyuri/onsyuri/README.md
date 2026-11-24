## Notes

[ONScripter-Yuri](https://github.com/YuriSizuku/OnscripterYuri) by YuriSizuku is an enhanced ONScripter (based on ONScripter-Jh) for NScripter visual novels. It is GPLv2; no games are included.

Copy each game into its own folder in `ports/onsyuri/games/`, e.g. `games/MyGame/` with `0.txt`, `00.txt` or `nscript.dat` and the `arc.nsa`/`.sar` archives. Saves are written to the game folder.

A launcher opens first: pick the game, the script encoding (Shift-JIS for Japanese and most English releases, GBK for Chinese, or UTF-8) and keep aspect ratio or stretch. Choose Play, or wait for it to start the last game.

Games without their own `default.ttf` use the bundled [WenQuanYi Micro Hei](http://wenq.org/) font (Latin, Chinese and Japanese).

## Controls

| Button | Action |
|--|--|
| A | Advance / select |
| B | Menu / cancel |
| X | Skip (hold) |
| Y | Space |
| Start | Auto mode |
| Select | Text speed |
| L1 | Skip to end of page |
| R1 | Skip mode |
| D-pad | Move between choices; Left = text history |
| Select + Start | Quit |

The engine reads the controller itself (upstream mapping).

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [YuriSizuku/OnscripterYuri](https://github.com/YuriSizuku/OnscripterYuri). The recipe folder [recipes/ports/onsyuri](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/onsyuri) holds `build.sh`, which builds a static Lua 5.4.7 and links Lua, libjpeg and bzip2 statically, then runs:

```sh
cmake -B build -DCMAKE_BUILD_TYPE=MinSizeRel && cmake --build build -j$(nproc)
```
