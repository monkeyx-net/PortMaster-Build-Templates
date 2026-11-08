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

GAMEBINARY=Nookery
GAMEDIR=/$directory/ports/nookery
CONFDIR="$GAMEDIR/conf"
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

cd $GAMEDIR
mkdir -p "$CONFDIR"

# The game needs the user's own disc image
disc=""
for f in "$GAMEDIR"/*.rvz "$GAMEDIR"/*.iso "$GAMEDIR"/*.gcm "$GAMEDIR"/*.ciso; do
  [ -f "$f" ] && disc="$f" && break
done
if [ -z "$disc" ]; then
  pm_message "Place your Animal Crossing (GAFE01) disc image in ports/nookery"
  sleep 5
  pm_finish
  exit 1
fi

# First run: OpenGL ES renderer, fullscreen, no prelaunch menu, no update check
[ -f "$CONFDIR/config.json" ] || cp "$GAMEDIR/config.json.default" "$CONFDIR/config.json"

$ESUDO chmod +x "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"
export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"
export game_libs="$GAMEDIR/libs.${DEVICE_ARCH}:$LD_LIBRARY_PATH"

# Aurora renders through Dawn (WebGPU), which only creates surfaces on X11/Wayland windows,
# so run it under Westonpack with the device's native EGL/GLES (crusty_x11egl).
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

# The game reads the gamepad itself (SDL3); gptokeyb only provides the exit hotkey
$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" &
pm_platform_helper "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"

# westonwrap evals the command line, so quote the disc path (names like "Animal Crossing (USA).iso")
$ESUDO env WRAPPED_LIBRARY_PATH=$game_libs \
$weston_dir/westonwrap.sh headless noop kiosk crusty_x11egl \
HOME=$CONFDIR SDL_VIDEO_DRIVER=x11 \
$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH} --user-dir "$CONFDIR" --disc "$(printf '%q' "$disc")"

$ESUDO $weston_dir/westonwrap.sh cleanup
if [[ "$PM_CAN_MOUNT" != "N" ]]; then
  $ESUDO umount "${weston_dir}"
fi
pm_finish
