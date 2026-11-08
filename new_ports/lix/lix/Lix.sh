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

GAMEDIR="/$directory/ports/lix"

cd "$GAMEDIR"
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

$ESUDO chmod +x "$GAMEDIR/lix.${DEVICE_ARCH}"

export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"

# Launcher (Love2D): player name (in user/options.sdl), cursor size, mouse speed and
# fullscreen mode (in lix.cfg), and the controls image
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/lix.cfg" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# Game files ship as assets.tar.gz; extract on first run and after an update.
# levels/ is extracted over, not removed, so levels saved by the editor are kept.
# extract/ (shared Love2D include) shows the message while it extracts and closes when
# done; it also shows the error if assets.tar.gz is missing or extraction fails.
if [[ -f assets.tar.gz || ! -d data ]]; then
    [[ -d data ]] && $ESUDO rm -rf data images music
    $LOVE_RUN "$GAMEDIR/extract" assets.tar.gz || exit 1
    $ESUDO rm -f assets.tar.gz
fi

# Launcher settings: minimum cursor scale (Lix default is 1), gptokeyb mouse_scale
# (lower is faster) and hardware (2, default) or software (1) fullscreen
source "$GAMEDIR/lix.cfg"
export LIX_CURSOR_MIN_SCALE
sed -i "s/^mouse_scale = .*/mouse_scale = $MOUSE_SCALE/" "$GAMEDIR/lix.gptk"
if [ "$SCREEN_MODE" = "1" ]; then
  FULLSCREEN="--fullscreen"
else
  FULLSCREEN="--hardfull=${DISPLAY_WIDTH}x${DISPLAY_HEIGHT}"
fi

# Bundled libs only for Lix: their libfreetype is older than some CFWs' and breaks
# the launcher's system libcairo (dArkOS: undefined symbol FT_Get_Transform)
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}:$LD_LIBRARY_PATH"

# Lix is mouse-driven: gptokeyb turns the left stick into a mouse
$GPTOKEYB "lix.${DEVICE_ARCH}" -c "$GAMEDIR/lix.gptk" &
pm_platform_helper "$GAMEDIR/lix.${DEVICE_ARCH}"
./lix.${DEVICE_ARCH} "$FULLSCREEN"

pm_finish
