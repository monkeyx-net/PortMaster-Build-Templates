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
GAMEDIR="/$directory/ports/starship"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/Starship.${DEVICE_ARCH}" "$GAMEDIR/assets/torch.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/mods"

# -------------------- BEGIN FUNCTIONS --------------------

# torch reads config.yml and assets/yaml/ from the working directory
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

# Starship's in-game extraction needs a desktop file dialog, so convert the roms with the
# standalone torch instead: US -> sf64.o2r, JP/EU/CN -> mods/sf64*.o2r voice packs
find_roms() {
    ROMS=()
    if [ ! -f "sf64.o2r" ]; then
        for rom in "$GAMEDIR/baseroms/"*.z64; do
            [ -f "$rom" ] && ROMS+=("$rom")
        done
        if [ ${#ROMS[@]} -eq 0 ]; then
            pm_message "Missing ROM! Place a Star Fox 64 US .z64 in baseroms/"
            sleep 5
            exit 1
        fi
    fi
}

extract_roms() {
    [ ${#ROMS[@]} -eq 0 ] && return
    for rom in "${ROMS[@]}"; do
        pm_message "Converting $(basename "$rom"), this can take a few minutes..."
        "$GAMEDIR/assets/torch.${DEVICE_ARCH}" o2r "$rom" || echo "torch failed on $rom"
    done
    if [ ! -f "sf64.o2r" ]; then
        pm_message "No sf64.o2r created! A Star Fox 64 US v1.0 or v1.1 rom is required."
        sleep 5
        exit 1
    fi
}

# Temporary swap for rom extraction on <=1GB devices: zram first (fast, any filesystem),
# swapfile/loop device only where zram isn't available
ZRAM_MB=1024
SWAP_MB=1024

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

    SWAPFILE="$dir/starship.swap"
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

# zram through sysfs, for systems without zramctl (BusyBox)
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

# zram via zramctl; adds a device rather than tearing down the system's own zram swap
swap_try_zramctl() {
    command -v zramctl >/dev/null 2>&1 || { echo "Swap: zramctl not found"; return 1; }
    dev=$($ESUDO zramctl --find --size "${ZRAM_MB}M" --algorithm lz4 2>/dev/null) \
         || dev=$($ESUDO zramctl --find --size "${ZRAM_MB}M" 2>/dev/null)
    if [ -z "$dev" ]; then
        echo "Swap: zramctl could not create a device"
        return 1
    fi
    ZRAMDEV="${dev#/dev/}"
    [ "$ZRAMDEV" != "zram0" ] && ZRAM_ADDED=1

    if $ESUDO mkswap "$dev" >/dev/null 2>&1 && $ESUDO swapon -p 5 "$dev"; then
        SWAP_DEV="$dev"
        echo "Swap enabled on $dev (zram ${ZRAM_MB}MB)"
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

# Swapfile fallback: Linux filesystems first (direct, else loop device), then exFAT/vfat (loop only)
swap_try_files() {
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
                    swap_try_dir "$dir" direct && return 0 ;;
                fat:exfat|fat:vfat)
                    tried="$tried$fsdev "
                    swap_try_dir "$dir" && return 0 ;;
                linux:*) echo "Swap: skipping $dir ($fstype)" ;;
            esac
        done
    done
    return 1
}

swap_on() {
    [ ${#ROMS[@]} -eq 0 ] && return
    [ "${DEVICE_RAM:-0}" -le 1 ] || return

    # Leave an existing big-enough zram swap (e.g. the CFW's own) alone; /proc/swaps lists sizes
    # a page short, so allow 1MB slack
    have_kb=$(awk '$1 ~ /zram/ { s += $3 } END { print s + 0 }' /proc/swaps)
    if [ $((have_kb + 1024)) -ge $((ZRAM_MB * 1024)) ]; then
        echo "Swap: zram swap already $((have_kb / 1024))MB, nothing to do"
        return
    fi

    swap_try_zramctl || swap_try_zram || swap_try_files || echo "Swap: none available, continuing without"
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
find_roms
swap_on
trap swap_off EXIT
extract_roms
swap_off

# Run the game
$GPTOKEYB "Starship.${DEVICE_ARCH}" -c "starship.gptk" &
pm_platform_helper "Starship.${DEVICE_ARCH}" >/dev/null
./Starship.${DEVICE_ARCH}

# Cleanup
rm -rf "$GAMEDIR/logs/"
pm_finish
