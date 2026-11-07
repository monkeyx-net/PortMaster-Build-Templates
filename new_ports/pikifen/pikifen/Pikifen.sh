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

GAMEDIR="/$directory/ports/pikifen"

# Pikifen keeps options and saves in user_data/ in the working directory
cd "$GAMEDIR"
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

$ESUDO chmod +x "$GAMEDIR/pikifen.${DEVICE_ARCH}"

# First run: key binds that match pikifen.gptk (type switching on Z/X
# instead of the mouse wheel, Space also confirms in menus)
if [ ! -f "$GAMEDIR/user_data/options.txt" ]; then
  mkdir -p "$GAMEDIR/user_data"
  cp "$GAMEDIR/conf/options.txt" "$GAMEDIR/user_data/options.txt"
fi

export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}:$LD_LIBRARY_PATH"
export SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig"

# Launcher (Love2D): mouse speed, fullscreen mode and frame rate, saved in
# pikifen.cfg, and the controls image
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/pikifen.cfg" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# Launcher settings: gptokeyb mouse_scale (lower is faster), hardware (1) or
# software (0) fullscreen and the game's frame rate. Always fullscreen at the
# device resolution.
source "$GAMEDIR/pikifen.cfg"
sed -i "s/^mouse_scale = .*/mouse_scale = $MOUSE_SCALE/" "$GAMEDIR/pikifen.gptk"
sed -i -e "s/^\tfullscreen=.*/\tfullscreen=true/" \
  -e "s/^\tresolution=.*/\tresolution=${DISPLAY_WIDTH} ${DISPLAY_HEIGHT}/" \
  -e "s/^\ttrue_fullscreen=.*/\ttrue_fullscreen=$([ "$SCREEN_MODE" = 1 ] && echo true || echo false)/" \
  -e "s/^\tfps=.*/\tfps=$FPS/" "$GAMEDIR/user_data/options.txt"

# Gamepad to keyboard and mouse: Pikifen has no default controller binds
$GPTOKEYB "pikifen.${DEVICE_ARCH}" -c "$GAMEDIR/pikifen.gptk" &
pm_platform_helper "$GAMEDIR/pikifen.${DEVICE_ARCH}"
./pikifen.${DEVICE_ARCH}

pm_finish
