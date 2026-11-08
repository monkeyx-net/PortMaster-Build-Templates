#!/bin/bash

XDG_DATA_HOME=${XDG_DATA_HOME:-$HOME/.local/share}

if [ -d "/opt/system/Tools/PortMaster/" ]; then
  controlfolder="/opt/system/Tools/PortMaster"
elif [ -d "/opt/tools/PortMaster/" ]; then
  controlfolder="/opt/tools/PortMaster"
elif [ -d "$XDG_DATA_HOME/PortMaster/" ]; then
  controlfolder="$XDG_DATA_HOME/PortMaster"
else
  controlfolder="/roms/ports/PortMaster"
fi

source $controlfolder/control.txt

[ -f "${controlfolder}/mod_${CFW_NAME}.txt" ] && source "${controlfolder}/mod_${CFW_NAME}.txt"
get_controls

GAMEBINARY=Nookery
GAMEDIR=/$directory/ports/nookery
CONFDIR="$GAMEDIR/conf"
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

cd $GAMEDIR
mkdir -p "$CONFDIR"

# The game needs the user's own disc image
disc=""
for f in "$GAMEDIR"/*.rvz "$GAMEDIR"/*.iso "$GAMEDIR"/*.gcm "$GAMEDIR"/*.ciso; do
  [ -f "$f" ] && disc="$f" && break
done
if [ -z "$disc" ]; then
  pm_message "Place your Animal Crossing (GAFE01) disc image in ports/nookery"
  sleep 5
  pm_finish
  exit 1
fi

# First run: OpenGL ES renderer, fullscreen, no prelaunch menu, no update check
[ -f "$CONFDIR/config.json" ] || cp "$GAMEDIR/config.json.default" "$CONFDIR/config.json"

# Launcher (Love2D): frame rate (config.json), aspect and render resolution (conf/launcher.cfg)
rm -f "$GAMEDIR/launch.cfg"
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$CONFDIR/config.json" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi
NOOKERY_ASPECT=auto
NOOKERY_RENDER_SCALE=0
[ -f "$CONFDIR/launcher.cfg" ] && source "$CONFDIR/launcher.cfg"

$ESUDO chmod +x "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"
export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"

# The game reads the gamepad itself (SDL3); gptokeyb only provides the exit hotkey
$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" &
pm_platform_helper "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"
# Dawn (WebGPU) cannot open a KMSDRM screen: AURORA_SDL2_DISPLAY=1 opens it with the CFW's SDL2.
# AURORA_NO_DRAW_MERGE=1: libmali only draws the first primitive of merged draws (dialogue text).
HOME="$CONFDIR" AURORA_SDL2_DISPLAY=1 AURORA_NO_DRAW_MERGE=1 NOOKERY_ASPECT=$NOOKERY_ASPECT NOOKERY_RENDER_SCALE=$NOOKERY_RENDER_SCALE \
./$GAMEBINARY.${DEVICE_ARCH} --user-dir "$CONFDIR" --disc "$disc"

pm_finish
