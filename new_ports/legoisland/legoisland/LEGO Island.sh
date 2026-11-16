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

# Set variables
GAMEDIR="/$directory/ports/legoisland"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/isle.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/saves"

# Launcher (Love2D): renderer, detail, transitions, frame rate, music; also sets the game paths
# in isle.ini
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/isle.ini" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# The game files are not included (see README.md): the whole CD in LEGO/, or its Lego/Scripts and
# DATA/DISK/LEGO/data folders as LEGO/Scripts and LEGO/data. cdpath and diskpath are the folders
# above them.
SCRIPTS=$(find "$GAMEDIR/LEGO" -maxdepth 4 -ipath '*/lego/scripts/credits.si' | head -n 1)
DATA=$(find "$GAMEDIR/LEGO" -maxdepth 6 -ipath '*/lego/data/world.wdb' | head -n 1)
if [ -z "$SCRIPTS" ] || [ -z "$DATA" ]; then
  pm_message "Missing game files! Copy your LEGO Island CD to ports/legoisland/LEGO"
  sleep 5
  pm_finish
  exit 0
fi
sed -i -e "s|^cdpath *=.*|cdpath = \"${SCRIPTS%/*/*/*}/\"|" \
  -e "s|^diskpath *=.*|diskpath = \"${DATA%/*/*/*}/\"|" \
  -e "s|^mediapath *=.*|mediapath = \"${DATA%/*/*/*}/\"|" "$GAMEDIR/isle.ini"

# libSDL3 is the SDL3-on-SDL2 shim: hand the CFW's driver choice to SDL2 and run SDL3 on it
if [ -n "$SDL_VIDEODRIVER" ]; then
  export SDL3SHIM_SDL2_VIDEODRIVER="$SDL_VIDEODRIVER" SDL_VIDEODRIVER=sdl2
fi
if [ -n "$SDL_AUDIODRIVER" ]; then
  export SDL3SHIM_SDL2_AUDIODRIVER="$SDL_AUDIODRIVER" SDL_AUDIODRIVER=sdl2
fi

# Run the game
$GPTOKEYB "isle.${DEVICE_ARCH}" -c "isle.gptk" &
pm_platform_helper "isle.${DEVICE_ARCH}" >/dev/null
./isle.${DEVICE_ARCH} --ini "$GAMEDIR/isle.ini"

pm_finish
