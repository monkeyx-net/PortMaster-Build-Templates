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
ZRAM_MB=512
ARKOS_ZRAM_MB=300

# Filesystem type of the mount holding $1 (BusyBox stat has no -f)
fs_type() {
    dir=$(realpath "$1" 2>/dev/null || echo "$1")
    awk -v d="$dir" '
        { mp = $2; gsub(/\\040/, " ", mp) }
        (d == mp || index(d, mp == "/" ? "/" : mp "/") == 1) && length(mp) > len { len = length(mp); t = $3 }
        END { print t }' /proc/mounts
}

# Try a swapfile in $1: direct swapon first (Linux filesystems), then through a loop device,
# which works where the kernel refuses swapfiles (btrfs before 5.0, exFAT/vfat)
swap_try_dir() {
    dir="$1"; direct="$2"
    free_mb=$(df -Pk "$dir" 2>/dev/null | awk 'NR==2 { print int($4 / 1024) }')
    if [ "${free_mb:-0}" -lt $((SWAP_MB + 100)) ]; then
        echo "Swap: skipping $dir (${free_mb:-0}MB free)"
        return 1
    fi

    SWAPFILE="$dir/soh.swap"
    pm_message "Low memory device, creating temporary swap..."
    if ! { $ESUDO dd if=/dev/zero of="$SWAPFILE" bs=1M count=$SWAP_MB >/dev/null 2>&1 \
           && $ESUDO chmod 600 "$SWAPFILE"; }; then
        echo "Swap: could not create $SWAPFILE"
        $ESUDO rm -f "$SWAPFILE"; SWAPFILE=
        return 1
    fi

    if [ -n "$direct" ] && $ESUDO mkswap "$SWAPFILE" >/dev/null 2>&1 \
       && $ESUDO swapon "$SWAPFILE" 2>/dev/null; then
        SWAP_DEV="$SWAPFILE"
        echo "Swap enabled at $SWAPFILE ($fstype)"
        return 0
    fi

    LOOPDEV=$($ESUDO losetup -f 2>/dev/null)
    if [ -n "$LOOPDEV" ] && $ESUDO losetup "$LOOPDEV" "$SWAPFILE" \
       && $ESUDO mkswap "$LOOPDEV" >/dev/null 2>&1 \
       && $ESUDO swapon "$LOOPDEV"; then
        SWAP_DEV="$LOOPDEV"
        echo "Swap enabled at $SWAPFILE via $LOOPDEV ($fstype)"
        return 0
    fi

    echo "Swap: failed at $SWAPFILE ($fstype)"
    [ -n "$LOOPDEV" ] && $ESUDO losetup -d "$LOOPDEV" 2>/dev/null
    $ESUDO rm -f "$SWAPFILE"
    SWAPFILE= LOOPDEV=
    return 1
}

# Last resort: compressed RAM swap. Helps less (the o2r buffer is already compressed) but is better than nothing
swap_try_zram() {
    [ -e /sys/block/zram0 ] || $ESUDO modprobe zram num_devices=1 2>/dev/null
    if [ -e /sys/block/zram0 ] && [ "$(cat /sys/block/zram0/disksize)" = "0" ]; then
        ZRAMDEV=zram0
    elif [ -e /sys/class/zram-control/hot_add ]; then
        ZRAMDEV="zram$($ESUDO cat /sys/class/zram-control/hot_add 2>/dev/null)"
        ZRAM_ADDED=1
    fi
    if [ -z "$ZRAMDEV" ] || [ ! -e "/sys/block/$ZRAMDEV" ]; then
        echo "Swap: zram not available"
        ZRAMDEV= ZRAM_ADDED=
        return 1
    fi

    echo lz4 | $ESUDO tee "/sys/block/$ZRAMDEV/comp_algorithm" >/dev/null 2>&1
    if echo "${ZRAM_MB}M" | $ESUDO tee "/sys/block/$ZRAMDEV/disksize" >/dev/null \
       && $ESUDO mkswap "/dev/$ZRAMDEV" >/dev/null 2>&1 \
       && $ESUDO swapon -p 5 "/dev/$ZRAMDEV"; then
        SWAP_DEV="/dev/$ZRAMDEV"
        echo "Swap enabled on /dev/$ZRAMDEV (zram ${ZRAM_MB}MB)"
        return 0
    fi

    echo "Swap: zram setup failed"
    zram_reset
    return 1
}

# ArkOS/dArkOS: zram via zramctl. Leaves an existing big-enough zram swap alone and adds a device
# rather than tearing down the system's own one
swap_try_zramctl() {
    command -v zramctl >/dev/null 2>&1 || { echo "Swap: zramctl not found"; return 1; }
    have_kb=$(awk '$1 ~ /zram/ { s += $3 } END { print s + 0 }' /proc/swaps)
    if [ "$have_kb" -ge $((ARKOS_ZRAM_MB * 1024)) ]; then
        echo "Swap: zram swap already $((have_kb / 1024))MB, nothing to do"
        return 0
    fi

    dev=$($ESUDO zramctl --find --size "${ARKOS_ZRAM_MB}M" 2>/dev/null)
    if [ -z "$dev" ]; then
        echo "Swap: zramctl could not create a device"
        return 1
    fi
    ZRAMDEV="${dev#/dev/}"
    [ "$ZRAMDEV" != "zram0" ] && ZRAM_ADDED=1

    if $ESUDO mkswap "$dev" >/dev/null 2>&1 && $ESUDO swapon -p 5 "$dev"; then
        SWAP_DEV="$dev"
        echo "Swap enabled on $dev (zram ${ARKOS_ZRAM_MB}MB)"
        return 0
    fi

    echo "Swap: zram setup failed on $dev"
    zram_reset
    return 1
}

zram_reset() {
    [ -n "$ZRAMDEV" ] || return
    echo 1 | $ESUDO tee "/sys/block/$ZRAMDEV/reset" >/dev/null 2>&1
    if [ -n "$ZRAM_ADDED" ]; then
        echo "${ZRAMDEV#zram}" | $ESUDO tee /sys/class/zram-control/hot_remove >/dev/null 2>&1
    fi
    ZRAMDEV= ZRAM_ADDED=
}

swap_on() {
    [ ${#ROMS[@]} -eq 0 ] && return
    [ "${DEVICE_RAM:-0}" -le 1 ] || return

    # ArkOS/dArkOS (btrfs root on kernel 4.4 refuses swapfiles): use zram
    if [[ "$CFW_NAME" == *"ArkOS"* ]]; then
        swap_try_zramctl || swap_try_zram || echo "Swap: none available, continuing without"
        [ -n "$SWAP_DEV" ] && cat /proc/swaps
        return
    fi

    # Prefer Linux filesystems (direct swapfile, else loop), then exFAT/vfat (loop only), then zram
    tried=" "
    for pass in linux fat; do
        for dir in "$GAMEDIR" /storage /userdata "$HOME" /var/tmp /opt /; do
            [ -d "$dir" ] && [ -w "$dir" -o -n "$ESUDO" ] || continue
            # several candidates often share one filesystem; only try each once
            fsdev=$(df -P "$dir" 2>/dev/null | awk 'NR==2 { print $1 }')
            case "$tried" in *" $fsdev "*) continue ;; esac
            fstype=$(fs_type "$dir")
            case "$pass:$fstype" in
                linux:ext2|linux:ext3|linux:ext4|linux:btrfs|linux:xfs|linux:f2fs)
                    tried="$tried$fsdev "
                    swap_try_dir "$dir" direct && break 2 ;;
                fat:exfat|fat:vfat)
                    tried="$tried$fsdev "
                    swap_try_dir "$dir" && break 2 ;;
                linux:*) echo "Swap: skipping $dir ($fstype)" ;;
            esac
        done
    done

    [ -n "$SWAP_DEV" ] || swap_try_zram || echo "Swap: none available, continuing without"
    [ -n "$SWAP_DEV" ] && cat /proc/swaps
}

swap_off() {
    [ -n "$SWAP_DEV" ] && $ESUDO swapoff "$SWAP_DEV" 2>/dev/null
    [ -n "$LOOPDEV" ] && $ESUDO losetup -d "$LOOPDEV" 2>/dev/null
    [ -n "$SWAPFILE" ] && $ESUDO rm -f "$SWAPFILE"
    zram_reset
    SWAP_DEV= LOOPDEV= SWAPFILE=
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

# torch writes the finished o2r to $TMPDIR before copying it here; /tmp is RAM (tmpfs) on most CFWs
if [ ${#ROMS[@]} -gt 0 ]; then
    rm -rf "$GAMEDIR/tmp"
    mkdir -p "$GAMEDIR/tmp"
    export TMPDIR="$GAMEDIR/tmp"
    pm_message "After extraction choose No at 'Run SoH?', the game restarts automatically"
fi

# Run the game
$GPTOKEYB "soh.elf.${DEVICE_ARCH}" -c "soh.gptk" &
pm_platform_helper "soh.elf.${DEVICE_ARCH}" >/dev/null
./soh.elf.${DEVICE_ARCH} "${ROMS[@]}"
status=$?

# torch keeps the whole o2r in RAM after extracting (upstream leak), so the game can lock up or crash
# in the same process; start it fresh if extraction just finished or that run died
if [ ${#ROMS[@]} -gt 0 ]; then
    swap_off
    rm -rf "$GAMEDIR/tmp"
    unset TMPDIR
    newest=$(ls -t oot.o2r oot-mq.o2r 2>/dev/null | head -n 1)
    if [ -n "$newest" ]; then
        age=$(( $(date +%s) - $(date -r "$newest" +%s) ))
        if [ $status -ne 0 ] || [ $age -lt 120 ]; then
            echo "Restarting SoH after extraction (exit $status, $newest ${age}s old)"
            ./soh.elf.${DEVICE_ARCH}
        fi
    fi
fi

# Cleanup
swap_off
rm -rf "$GAMEDIR/tmp"
rm -rf "$GAMEDIR/logs/"
pm_finish
