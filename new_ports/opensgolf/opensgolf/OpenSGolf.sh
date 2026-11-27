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

GAMEBINARY=opensgolf
GAMEDIR=/$directory/ports/opensgolf
CONFDIR="$GAMEDIR/conf"
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

cd $GAMEDIR
mkdir -p "$CONFDIR"
chmod +x "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"

# Optional: a copy of the original SimGolf (the folder with Data, Flics and Sounds) in simgolf/.
# Without it the game uses its built-in free graphics and sounds.
GAMEARGS="--size ${DISPLAY_WIDTH}x${DISPLAY_HEIGHT}"
[ -d "$GAMEDIR/simgolf" ] && GAMEARGS="$GAMEARGS --game $GAMEDIR/simgolf"

# OpenSGolf uses miniquad (X11 + OpenGL), so run it under Westonpack.
weston_dir=/tmp/weston
$ESUDO mkdir -p "${weston_dir}"
weston_runtime="weston_pkg_0.2"
if [ ! -f "$controlfolder/libs/${weston_runtime}.squashfs" ]; then
  if [ ! -f "$controlfolder/harbourmaster" ]; then
    pm_message "This port requires the latest PortMaster to run, please go to https://portmaster.games/ for more info."
    sleep 5
    exit 1
  fi
  $ESUDO $controlfolder/harbourmaster --quiet --no-check runtime_check "${weston_runtime}.squashfs"
fi
if [[ "$PM_CAN_MOUNT" != "N" ]]; then
  $ESUDO umount "${weston_dir}"
fi
$ESUDO mount "$controlfolder/libs/${weston_runtime}.squashfs" "${weston_dir}"

$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" -c "./$GAMEBINARY.gptk" &
pm_platform_helper "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"

# CRUSTY_SHOW_CURSOR: the game uses the system mouse cursor, which crusty only draws when asked.
$ESUDO env CRUSTY_BLOCK_INPUT=1 CRUSTY_SHOW_CURSOR=1 \
$weston_dir/westonwrap.sh headless noop kiosk crusty_glx_gl4es \
HOME=$CONFDIR XDG_DATA_HOME=$CONFDIR WAYLAND_DISPLAY= \
$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH} $GAMEARGS

$ESUDO $weston_dir/westonwrap.sh cleanup
if [[ "$PM_CAN_MOUNT" != "N" ]]; then
  $ESUDO umount "${weston_dir}"
fi
pm_finish
