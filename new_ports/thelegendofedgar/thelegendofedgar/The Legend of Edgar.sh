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

GAMEBINARY=edgar
GAMEDIR=/$directory/ports/edgar
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1


bind_directories ~/.parallelrealities/edgar $GAMEDIR/conf

cd $GAMEDIR

source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN controls/controls.love controls/controls.png 30 $DISPLAY_WIDTH $DISPLAY_HEIGHT

$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" -c "./$GAMEBINARY.gptk" &
SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig" ./$GAMEBINARY.${DEVICE_ARCH} -nojoystick
pm_finish
