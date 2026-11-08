# Open Golf

## Notes

A 3D minigolf game with 20 holes. Thanks to [Michael Gerdes](https://github.com/mgerdes/Open-Golf) for creating the game (MIT). Models are from [Kenney](https://kenney.nl/assets) (CC0).

This port builds the `portmaster` branch of [monkeyx-net/Open-Golf](https://github.com/monkeyx-net/Open-Golf/tree/portmaster), which runs the game on SDL2 with OpenGL ES 3.0 and adds gamepad controls. The menus use an on-screen pointer. To putt, hold A (the pointer jumps to the ball), pull back with the stick to set the direction and power, and release A.

Progress is saved in `storage.json` in the port folder.

## Controls

| Button | Action |
|--|--|
| A | Aim/OK |
| B | Back |
| L1 | Cam ← |
| R1 | Cam → |
| L2 | Zoom in |
| R2 | Zoom out |
| Start | Pause |
| Select | Exit (+Start) |
| D-pad | Pointer (slow) |
| Left stick | Pointer / aim |
| Right stick | Camera / scroll |

Hold A to grab the ball, pull back with the left stick, and release A to putt.

## Compiling

`cmake -S . -B cmake-build -DGOLF_PORTMASTER=ON -DCMAKE_BUILD_TYPE=Release && cmake --build cmake-build`
