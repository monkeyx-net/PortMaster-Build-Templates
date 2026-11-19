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
GAMEDIR="/$directory/ports/soh2"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/2s2h.elf.${DEVICE_ARCH}"

# -------------------- BEGIN FUNCTIONS --------------------

# 2s2h converts roms itself (ZAPD is built in) using Config_*.xml, filelists/ and xml/ in assets/
unzip_assets() {
    if [ -f "$GAMEDIR/assets/extractor.zip" ]; then
        $LOVE_RUN "$GAMEDIR/extract" --text "Unpacking extractor assets..." --dir "$GAMEDIR/assets" assets/extractor.zip || exit 1
        rm -f "$GAMEDIR/assets/extractor.zip"
    fi
}

# Pass baseroms to 2s2h.elf when no o2r exists yet; it converts them and then starts the game
otr_check() {
    ROMS=()
    if [ ! -f "mm.o2r" ]; then
        for rom in "$GAMEDIR/baseroms/"*.z64 "$GAMEDIR/baseroms/"*.n64 "$GAMEDIR/baseroms/"*.v64; do
            [ -f "$rom" ] && ROMS+=("$rom")
        done
        if [ ${#ROMS[@]} -eq 0 ]; then
            pm_message "Missing ROM files! Place a .z64, .n64 or .v64 in baseroms/"
            sleep 5
            pm_finish
            exit 0
        fi
    fi
}

# --------------------- END FUNCTIONS ---------------------

# Launcher (Love2D): frame rate and N64 low-res mode; also closes the game menu and enables
# controller navigation in 2ship2harkinian.json
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/2ship2harkinian.json" "$GAMEDIR/launch.cfg"

# Continue only if the user chose Play (launcher writes LAUNCHER_PLAY=1 to launch.cfg)
if ! grep -qx "LAUNCHER_PLAY=1" "$GAMEDIR/launch.cfg" 2>/dev/null; then
  pm_finish
  exit 0
fi

# Perform functions
unzip_assets
otr_check

# Run the game
$GPTOKEYB "2s2h.elf.${DEVICE_ARCH}" -c "soh2.gptk" &
pm_platform_helper "2s2h.elf.${DEVICE_ARCH}" >/dev/null
./2s2h.elf.${DEVICE_ARCH} "${ROMS[@]}"

# Cleanup
rm -rf "$GAMEDIR/logs/"
pm_finish
