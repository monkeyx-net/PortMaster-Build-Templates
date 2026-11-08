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

# First run: Lix asks for a player name, which needs a keyboard. Pre-seed one
# and default to fullscreen at the device's resolution.
if [ ! -f "$GAMEDIR/user/options.sdl" ]; then
  mkdir -p "$GAMEDIR/user"
  printf 'userName "Player"\nscreenMode 1\n' > "$GAMEDIR/user/options.sdl"
fi

export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}:$LD_LIBRARY_PATH"
export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"

# Launcher (Love2D): cursor size, mouse speed and fullscreen mode, saved in lix.cfg,
# and the controls image
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/lix.cfg" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# Game files ship as assets.tar.gz; extract on first run and after an update.
# Runs after the launcher, so the message window is not on screen with it.
# levels/ is extracted over, not removed, so levels saved by the editor are kept.
# The message is shown with PortMaster's portmsg run as this user, not pm_message:
# on sudo CFWs (ArkOS/dArkOS) pm_message's root copy can't be closed again.
# Older PortMaster without portmsg falls back to pm_message.
PORTMSG="$controlfolder/portmsg.${DEVICE_ARCH}"

ARCHIVE_FILE="assets.tar.gz"
if [[ -f "$ARCHIVE_FILE" ]]; then
    [[ -d data ]] && $ESUDO rm -rf data images music
    MSG="Extracting game data, this can take a few minutes..."
    if [ -x "$PORTMSG" ]; then
        "$PORTMSG" --wait --type info --text "$MSG" &
        MSG_PID=$!
    else
        pm_message "$MSG"
    fi
    gunzip -c "$ARCHIVE_FILE" | tar --no-same-owner -m -xf -
    EXTRACT_RC=$?
    [ -n "$MSG_PID" ] && { kill "$MSG_PID"; wait "$MSG_PID"; }
    if [ $EXTRACT_RC -ne 0 ]; then
        pm_show_error "Extraction failed. Check free space and launch again."
        exit 1
    fi
    $ESUDO rm -f "$ARCHIVE_FILE"
elif [[ ! -d data ]]; then
    pm_show_error "No game data and $ARCHIVE_FILE not found. Reinstall the port."
    exit 1
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

# Lix is mouse-driven: gptokeyb turns the left stick into a mouse
$GPTOKEYB "lix.${DEVICE_ARCH}" -c "$GAMEDIR/lix.gptk" &
pm_platform_helper "$GAMEDIR/lix.${DEVICE_ARCH}"
./lix.${DEVICE_ARCH} "$FULLSCREEN"

pm_finish
