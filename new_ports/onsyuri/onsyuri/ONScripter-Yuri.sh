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

GAMEDIR="/$directory/ports/onsyuri"

cd "$GAMEDIR"
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

$ESUDO chmod +x "$GAMEDIR/onsyuri.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/games"

export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"

# Launcher (Love2D): game folder in games/, script encoding and screen mode (in onsyuri.cfg)
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/onsyuri.cfg" "$GAMEDIR/launch.cfg"

if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

source "$GAMEDIR/onsyuri.cfg"
if [ -z "$GAME" ] || [ ! -d "$GAMEDIR/games/$GAME" ]; then
  pm_finish
  exit 0
fi

# Games without their own default.ttf use the bundled CJK font
ARGS=(--root "$GAMEDIR/games/$GAME" "--enc:$ENCODING")
[ -f "$GAMEDIR/games/$GAME/default.ttf" ] || ARGS+=(--font "$GAMEDIR/font/wqy-microhei.ttc")
if [ "$SCREEN" = "stretch" ]; then
  ARGS+=(--fullscreen2)
else
  ARGS+=(--fullscreen)
fi

# The engine reads the controller itself; gptokeyb only handles the exit hotkey
$GPTOKEYB "onsyuri.${DEVICE_ARCH}" &
pm_platform_helper "$GAMEDIR/onsyuri.${DEVICE_ARCH}"
./onsyuri.${DEVICE_ARCH} "${ARGS[@]}"

pm_finish
