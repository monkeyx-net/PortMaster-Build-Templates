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
GAMEDIR="/$directory/ports/paperboat"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/Paperboat.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/mods"

# -------------------- BEGIN FUNCTIONS --------------------

# The game's built-in torch reads config.yml and assets/ from its folder
unzip_assets() {
    if [ -f "$GAMEDIR/assets/extractor.zip" ]; then
        $LOVE_RUN "$GAMEDIR/extract" --text "Unpacking extractor assets..." assets/extractor.zip || exit 1
        rm -f "$GAMEDIR/assets/extractor.zip"
    fi
}

# The game converts a .z64 in baseroms/ to pm64.o2r on first run (it asks on screen first)
find_roms() {
    ROMS=()
    if [ ! -f "pm64.o2r" ]; then
        for rom in "$GAMEDIR/baseroms/"*.z64; do
            [ -f "$rom" ] && ROMS+=("$rom")
        done
        if [ ${#ROMS[@]} -eq 0 ]; then
            pm_message "Missing ROM! Place a Paper Mario US .z64 in ports/paperboat/baseroms/"
            sleep 5
            pm_finish
            exit 0
        fi
    fi
}

# Temporary swap for the rom conversion on <=1GB devices (swap_on/swap_off)
source "$GAMEDIR/swap.txt"

# --------------------- END FUNCTIONS ---------------------

# The settings launcher edits paperboat.cfg.json line by line, so seed a pretty-printed one;
# controller navigation lets the gamepad answer the rom conversion prompts
[ -f paperboat.cfg.json ] || printf '{\n    "CVars": {\n        "gSettings": {\n            "ControlNav": 1\n        }\n    }\n}\n' > paperboat.cfg.json

# Launcher (Love2D): frame rate and N64 low-res mode; also closes the menu bar and enables
# controller navigation in paperboat.cfg.json
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/paperboat.cfg.json" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# Perform functions
unzip_assets
find_roms
swap_on
trap swap_off EXIT

# Run the game
$GPTOKEYB "Paperboat.${DEVICE_ARCH}" -c "paperboat.gptk" &
pm_platform_helper "Paperboat.${DEVICE_ARCH}" >/dev/null
./Paperboat.${DEVICE_ARCH}

# Cleanup
swap_off
rm -rf "$GAMEDIR/logs/"
pm_finish
