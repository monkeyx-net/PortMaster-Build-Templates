## Notes
Thanks to [Martin Taibr](https://github.com/martin-t) for creating RecWars, a FOSS Rust remake of RecWar by Willem Janssen.

[https://github.com/martin-t/rec-wars](https://github.com/martin-t/rec-wars)

Runs under the Westonpack runtime (macroquad needs X11/OpenGL). The game starts directly in `local` mode against bots.

## Controls

| Button | Action |
|--|--|
| Dpad / Left Analog | Drive |
| Right Analog / L1 / R1 | Rotate turret left / right |
| A | Fire |
| B | Lay mine |
| X | Previous weapon |
| Y | Next weapon |
| L2 | Horn |
| R2 | Fire (alternate) |
| L3 / R3 | Self destruct |
| Start | Pause |
| Select | Esc |
| Select + Start | Quit |

## Compiling

GitHub Actions automatic build: `recipes/ports/recwars` (`cargo build --release --locked`).
