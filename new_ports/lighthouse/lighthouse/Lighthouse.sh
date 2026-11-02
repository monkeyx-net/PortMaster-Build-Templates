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
GAMEDIR="/$directory/ports/lighthouse"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/Lighthouse.${DEVICE_ARCH}"

# -------------------- BEGIN FUNCTIONS --------------------

# Check imgui.ini and modify if needed
imgui_reset() {
    input_file="imgui.ini"
    temp_file="imgui_temp.ini"
    skip_section=0
    # Loop through each line in the input file
    while IFS= read -r line; do
        # Check if the line is a window header
        if [[ "$line" =~ ^\[Window\]\[Main\ Game\] || "$line" =~ ^\[Window\]\[Main\ -\ Deck\] ]]; then
            skip_section=1  # Set the flag to skip modifications for this section
        elif [[ "$line" =~ ^\[Window\] ]]; then
            skip_section=0  # Reset the flag for other windows
        fi

        # Modify Pos and Size only if the current section is not skipped
        if [[ $skip_section -eq 0 ]]; then
            if [[ "$line" =~ ^Pos=.* ]]; then
                echo "Pos=30,30" >> "$temp_file"
            elif [[ "$line" =~ ^Size=.* ]]; then
                echo "Size=400,300" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        else
            # If skipping, write the line unchanged
            echo "$line" >> "$temp_file"
        fi
    done < "$input_file"

    # Replace the original file with the modified one
    mv "$temp_file" "$input_file"
}

# Lighthouse extracts roms itself (torch) using config.yml and assets/yaml/ from its folder
unzip_assets() {
    if [ -f "$GAMEDIR/assets/extractor.zip" ]; then
        pm_message "Unpacking extractor assets..."
        if "$controlfolder/7zzs.${DEVICE_ARCH}" x -y "$GAMEDIR/assets/extractor.zip" -o"$GAMEDIR" >/dev/null; then
            rm -f "$GAMEDIR/assets/extractor.zip"
        else
            pm_message "Failed to unpack extractor.zip!"
            sleep 5
            exit 1
        fi
    fi
}

# Pass baseroms to Lighthouse when no o2r exists yet; it prompts on screen and then starts the game
otr_check() {
    ROMS=()
    if [ ! -f "bk.o2r" ]; then
        for rom in "$GAMEDIR/baseroms/"*.z64; do
            [ -f "$rom" ] && ROMS+=("$rom")
        done
        if [ ${#ROMS[@]} -eq 0 ]; then
            pm_message "Missing ROM! Place a Banjo-Kazooie .z64 in baseroms/"
            sleep 5
            exit 1
        fi
    fi
}

# Temporary swap for rom extraction on <=1GB devices (swap_on/swap_off)
source "$GAMEDIR/swap.txt"

# --------------------- END FUNCTIONS ---------------------

# Launcher (Love2D): Anchor player name / room ID; also closes the game menu and
# enables controller navigation in lighthouse.cfg.json
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/lighthouse.cfg.json" "$GAMEDIR/launch.cfg"

LAUNCHER_PLAY=0
[ -f "$GAMEDIR/launch.cfg" ] && source "$GAMEDIR/launch.cfg"
if [ "$LAUNCHER_PLAY" != "1" ]; then
  echo "Launcher exited without starting the game"
  pm_finish
  exit 0
fi

# Perform functions
unzip_assets
otr_check
swap_on
trap swap_off EXIT

if [ -f "imgui.ini" ]; then
    imgui_reset
fi

# torch writes the finished o2r to $TMPDIR before copying it here; /tmp is RAM (tmpfs) on most CFWs
if [ ${#ROMS[@]} -gt 0 ]; then
    rm -rf "$GAMEDIR/tmp"
    mkdir -p "$GAMEDIR/tmp"
    export TMPDIR="$GAMEDIR/tmp"
    pm_message "After extraction choose No at 'Run Lighthouse?', the game restarts automatically"
fi

# Run the game
$GPTOKEYB "Lighthouse.${DEVICE_ARCH}" -c "lighthouse.gptk" &
pm_platform_helper "Lighthouse.${DEVICE_ARCH}" >/dev/null
./Lighthouse.${DEVICE_ARCH} "${ROMS[@]}"
status=$?

# After extraction Lighthouse offers to run the game in the same process; start it fresh
# instead if the user chose No (bk.o2r written under 2 minutes ago) or that run died
if [ ${#ROMS[@]} -gt 0 ]; then
    swap_off
    rm -rf "$GAMEDIR/tmp"
    unset TMPDIR
    if [ -f bk.o2r ] && { [ $status -ne 0 ] || [ $(( $(date +%s) - $(date -r bk.o2r +%s) )) -lt 120 ]; }; then
        echo "Restarting Lighthouse after extraction (exit $status)"
        ./Lighthouse.${DEVICE_ARCH}
    fi
fi

# Cleanup
swap_off
rm -rf "$GAMEDIR/tmp"
rm -rf "$GAMEDIR/logs/"
pm_finish
