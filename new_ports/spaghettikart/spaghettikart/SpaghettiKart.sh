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
GAMEDIR="/$directory/ports/spaghettikart"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs:$LD_LIBRARY_PATH"
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# Set up logging
BINARY="Spaghettify.${DEVICE_ARCH}"
cd $GAMEDIR
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

# Permissions
$ESUDO chmod +x "$GAMEDIR/$BINARY
$ESUDO chmod +x "$GAMEDIR/tools/otrgen"
$ESUDO chmod +x "$GAMEDIR/tools/torch"

# Close the menu if open
sed -i 's/"gOpenMenu": *1/"gOpenMenu": 0/' spaghettify.cfg.json

# Warn if mk64.o2r is older than Spaghettify or spaghetti.o2r
if [ -f "$GAMEDIR/mk64.o2r" ]; then
    if [ -f "$GAMEDIR/$BINARY" ] && [ "$GAMEDIR/$BINARY" -nt "$GAMEDIR/mk64.o2r" ] \
       || [ -f "$GAMEDIR/spaghetti.o2r" ] && [ "$GAMEDIR/spaghetti.o2r" -nt "$GAMEDIR/mk64.o2r" ]; then
        echo "Notice: mk64.o2r is older than Spaghettify and/or spaghetti.o2r. Forcing regeneration."
        rm -f "$GAMEDIR/mk64.o2r"
        REGEN=1
        export REGEN
    fi
fi

# Check if we need to generate any o2r files
if [ ! -f "$GAMEDIR/mk64.o2r" ]; then
    # Ensure we have a rom file before attempting to generate o2r
    if ls "$GAMEDIR/baseroms/"*.*64 1> /dev/null 2>&1; then
        if [ -f "$controlfolder/utils/patcher.txt" ]; then
            export PATCHER_FILE="$GAMEDIR/tools/otrgen"
            export PATCHER_GAME="$(basename "${0%.*}")"
            export PATCHER_TIME="5 to 10 minutes"
            export controlfolder
            source "$controlfolder/utils/patcher.txt"
            pid=$(pidof gptokeyb) && [ -n "$pid" ] && $ESUDO kill -9 $pid
        else
            pm_message "This port requires the latest version of PortMaster."
        fi
    else
        pm_message "Missing ROM files! Can't generate o2r!"
    fi
fi

# Check if O2R files were generated
if [ ! -f "$GAMEDIR/mk64.o2r" ]; then
    pm_message "No o2r found, can't run the game!"
    exit 1
fi

# Run the game
$GPTOKEYB "$BINARY" -c "spaghetti.gptk" &
pm_platform_helper "$GAMEDIR/$BINARY" > /dev/null
./$BINARY

# Cleanup
rm -rf logs
pm_finish
