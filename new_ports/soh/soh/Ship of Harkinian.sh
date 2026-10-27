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

# Temporary swap for rom extraction on <=1GB devices (torch builds the whole o2r in RAM)
SWAP_MB=1024

# Filesystem type of the mount holding $1 (BusyBox stat has no -f)
fs_type() {
    dir=$(realpath "$1" 2>/dev/null || echo "$1")
    awk -v d="$dir" '
        { mp = $2; gsub(/\\040/, " ", mp) }
        (d == mp || index(d, mp == "/" ? "/" : mp "/") == 1) && length(mp) > len { len = length(mp); t = $3 }
        END { print t }' /proc/mounts
}

swap_on() {
    [ ${#ROMS[@]} -eq 0 ] && return
    [ "${DEVICE_RAM:-0}" -le 1 ] || return

    # /roms is often exFAT, which can't hold a swapfile, so look for a Linux filesystem elsewhere
    for dir in "$GAMEDIR" /storage /userdata "$HOME" /var/tmp /opt /; do
        [ -d "$dir" ] && [ -w "$dir" -o -n "$ESUDO" ] || continue
        fstype=$(fs_type "$dir")
        case "$fstype" in
            ext2|ext3|ext4|btrfs|xfs|f2fs) ;;
            *) echo "Swap: skipping $dir ($fstype)"; continue ;;
        esac
        free_mb=$(df -Pk "$dir" 2>/dev/null | awk 'NR==2 { print int($4 / 1024) }')
        if [ "${free_mb:-0}" -lt $((SWAP_MB + 100)) ]; then
            echo "Swap: skipping $dir (${free_mb:-0}MB free)"
            continue
        fi
        SWAPFILE="$dir/soh.swap"
        break
    done

    if [ -z "$SWAPFILE" ]; then
        echo "Swap: no suitable filesystem found, continuing without"
        return
    fi
    grep -q "$SWAPFILE" /proc/swaps && return

    pm_message "Low memory device, creating temporary swap..."
    if $ESUDO dd if=/dev/zero of="$SWAPFILE" bs=1M count=$SWAP_MB >/dev/null 2>&1 \
       && $ESUDO chmod 600 "$SWAPFILE" \
       && $ESUDO mkswap "$SWAPFILE" >/dev/null 2>&1 \
       && $ESUDO swapon "$SWAPFILE"; then
        SWAP_ON=1
        echo "Swap enabled at $SWAPFILE ($fstype):"
        cat /proc/swaps
    else
        echo "Swap setup failed at $SWAPFILE, continuing without"
        $ESUDO rm -f "$SWAPFILE"
    fi
}

swap_off() {
    if [ -n "$SWAP_ON" ]; then
        $ESUDO swapoff "$SWAPFILE"
        $ESUDO rm -f "$SWAPFILE"
        SWAP_ON=
    fi
}

# --------------------- END FUNCTIONS ---------------------

# Perform functions
unzip_assets
otr_check
swap_on
trap swap_off EXIT

if [ -f "imgui.ini" ]; then
    imgui_reset
fi

# Run the game
$GPTOKEYB "soh.elf.${DEVICE_ARCH}" -c "soh.gptk" &
pm_platform_helper "soh.elf.${DEVICE_ARCH}" >/dev/null
./soh.elf.${DEVICE_ARCH} "${ROMS[@]}"

# Cleanup
swap_off
rm -rf "$GAMEDIR/logs/"
pm_finish
