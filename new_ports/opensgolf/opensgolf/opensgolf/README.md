## Notes

Thanks to [crspytopgn](https://github.com/crspytopgn/opensgolf) for OpenSGolf, an open source engine for Sid Meier's SimGolf (2002).

The game is ready to play with its built-in free graphics and sounds. To play with the original art, sounds and music, copy the folder of your own SimGolf copy (the one containing `Data`, `Flics` and `Sounds`) to `ports/opensgolf/simgolf/`. The game files are not included.

The game was made for 800x600, so text is small on 640x480 screens.

## Controls

| Button | Action |
|--|--|
| Left Analog | Move the pointer |
| Right Analog | Scroll the course |
| A | Click (hold to drag or paint) |
| B | Undo/remove on the course, back elsewhere |
| X | Turn building / next design |
| Y | Game speed |
| L1/R1 | Turn the view |
| L2/R2 | Zoom out/in |
| D-Pad | Arrow keys (scroll, lists) |
| Start | System Functions (save, load, quit) |
| Select | Pause |
| L3 | Open hole |
| R3 | Repeat last message |

## Compile

Built with [PortMaster-Build-Templates](https://github.com/monkeyx-net/PortMaster-Build-Templates) from [crspytopgn/opensgolf](https://github.com/crspytopgn/opensgolf). The recipe folder [recipes/ports/opensgolf](https://github.com/monkeyx-net/PortMaster-Build-Templates/tree/main/recipes/ports/opensgolf) holds `build.sh`, which reads `recipe.json` and applies the patches and files in that folder before running the build command:

```sh
cargo build --release --locked -p opensgolf --no-default-features --features audio,gamepad
```
