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
GAMEDIR="/$directory/ports/soh"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/soh.elf.${DEVICE_ARCH}"

# Close the menu if open
sed -i 's/"Menu": *1/"Menu": 0/' shipofharkinian.json

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

# SoH 9.3+ extracts roms itself (torch) using the ymls in assets/, so no patcher is needed
unzip_assets() {
    if [ -f "$GAMEDIR/assets/extractor.zip" ]; then
        pm_message "Unpacking extractor assets..."
        if "$controlfolder/7zzs.${DEVICE_ARCH}" x -y "$GAMEDIR/assets/extractor.zip" -o"$GAMEDIR/assets" >/dev/null; then
            rm -f "$GAMEDIR/assets/extractor.zip"
        else
            pm_message "Failed to unpack extractor.zip!"
            sleep 5
            exit 1
        fi
    fi
}

# Pass baseroms to soh.elf when no o2r exists yet; it prompts on screen and then starts the game
otr_check() {
    ROMS=()
    if [ ! -f "oot.o2r" ] && [ ! -f "oot-mq.o2r" ]; then
        for rom in "$GAMEDIR/baseroms/"*.z64 "$GAMEDIR/baseroms/"*.n64 "$GAMEDIR/baseroms/"*.v64; do
            [ -f "$rom" ] && ROMS+=("$rom")
        done
        if [ ${#ROMS[@]} -eq 0 ]; then
            pm_message "Missing ROM files! Place a .z64, .n64 or .v64 in baseroms/"
            sleep 5
            exit 1
        fi
    fi
}

# --------------------- END FUNCTIONS ---------------------

# Perform functions
unzip_assets
otr_check

if [ -f "imgui.ini" ]; then
    imgui_reset
fi

# Run the game
$GPTOKEYB "soh.elf.${DEVICE_ARCH}" -c "soh.gptk" &
pm_platform_helper "soh.elf.${DEVICE_ARCH}" >/dev/null
./soh.elf.${DEVICE_ARCH} "${ROMS[@]}"

# Cleanup
rm -rf "$GAMEDIR/logs/"
pm_finish
