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
GAMEDIR="/$directory/ports/paperboat"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/Paperboat.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/mods"

# First run: open with the menu bar closed and controller navigation on, so the
# rom conversion prompts can be answered with the gamepad
[ -f paperboat.cfg.json ] || echo '{"CVars":{"gSettings":{"ControlNav":1,"OpenMenuBar":0}}}' > paperboat.cfg.json

# The game converts a .z64 in its own folder to pm64.o2r on first run
ROMS=()
if [ ! -f "pm64.o2r" ]; then
  for rom in "$GAMEDIR/"*.z64; do
    [ -f "$rom" ] && ROMS+=("$rom")
  done
  if [ ${#ROMS[@]} -eq 0 ]; then
    pm_message "Missing ROM! Place a Paper Mario US .z64 in ports/paperboat/"
    sleep 5
    pm_finish
    exit 0
  fi
fi

# Temporary swap for the rom conversion on <=1GB devices (swap_on/swap_off)
source "$GAMEDIR/swap.txt"
swap_on
trap swap_off EXIT

# Run the game
$GPTOKEYB "Paperboat.${DEVICE_ARCH}" -c "paperboat.gptk" &
pm_platform_helper "Paperboat.${DEVICE_ARCH}" >/dev/null
./Paperboat.${DEVICE_ARCH}

# Cleanup
swap_off
rm -rf "$GAMEDIR/logs/"
pm_finish
