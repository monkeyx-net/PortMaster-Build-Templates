## Notes

Pikifen is a free, MIT-licensed fan engine and game inspired by Pikmin
(Pikmin is (c) Nintendo). Asset credits and licenses are in
`licenses/LICENSE.Pikifen-assets.txt`.
Thanks to André 'Espyo' Silva and the Pikifen contributors: https://pikminfanon.com/wiki/Pikifen

Built with Allegro 5 using its SDL2 + OpenGL ES 2 backend. The shaders were
converted to GLSL ES and the editors' file dialogs are disabled (no GTK on
handhelds).

## Launcher
A settings launcher opens before the game: mouse speed (Slow, Normal, Fast, Faster), hardware (default) or software fullscreen, frame rate (60 or 30, for slower devices), and the controls image. Choose Play, or wait for it to start the game.

## Controls

| Button | Action |
|--|--|
| A | Throw |
| B | Whistle |
| X | Swarm |
| Y | Inventory |
| L1 | Prev type |
| R1 | Next type |
| L2 | Dismiss |
| R2 | Next leader |
| R3 | Zoom |
| Start | Pause |
| Select | Radar |
| Left stick | Move |
| Right stick | Cursor |
| D-pad | Menu |

The gamepad is mapped to keyboard and mouse. A also clicks and X confirms in menus; Start goes back. In-game button hints show the keyboard keys. Binds can be changed in Pikifen's options; `pikifen.gptk` maps the gamepad to keys.
