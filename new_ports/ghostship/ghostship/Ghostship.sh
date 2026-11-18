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
GAMEDIR="/$directory/ports/ghostship"

# Exports
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}":$LD_LIBRARY_PATH
export SDL_GAMECONTROLLERCONFIG=$sdl_controllerconfig

# CD and set log
cd $GAMEDIR
: > "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1
$ESUDO chmod +x "$GAMEDIR/Ghostship.${DEVICE_ARCH}" "$GAMEDIR/assets/torch.${DEVICE_ARCH}"
mkdir -p "$GAMEDIR/mods"

# -------------------- BEGIN FUNCTIONS --------------------

# torch reads config.yml and assets/ymls from the working directory
unzip_assets() {
    if [ -f "$GAMEDIR/assets/extractor.zip" ]; then
        $LOVE_RUN "$GAMEDIR/extract" --text "Unpacking extractor assets..." assets/extractor.zip || exit 1
        rm -f "$GAMEDIR/assets/extractor.zip"
    fi
}

# sm64.o2r is built from the rom by the standalone torch, so the game never shows its own
# conversion prompts. Rebuild it when the port was updated after it was made (and the rom is still there).
find_roms() {
    ROMS=()
    for rom in "$GAMEDIR/baseroms/"*.z64; do
        [ -f "$rom" ] && ROMS+=("$rom")
    done
    if [ -f "sm64.o2r" ]; then
        if [ ${#ROMS[@]} -gt 0 ] && [ "Ghostship.${DEVICE_ARCH}" -nt "sm64.o2r" ]; then
            echo "sm64.o2r is older than the port, rebuilding it"
            rm -f "sm64.o2r"
        else
            ROMS=()
        fi
    elif [ ${#ROMS[@]} -eq 0 ]; then
        pm_message "Missing ROM! Place a Super Mario 64 US or JP .z64 in ports/ghostship/baseroms/"
        sleep 5
        pm_finish
        exit 0
    fi
}

extract_roms() {
    [ ${#ROMS[@]} -eq 0 ] && return
    # -u stores the port version in sm64.o2r, which the game compares with its own
    VERSION=$(sed -n 's/.*"version": *"v\{0,1\}\([0-9.]*\)".*/\1/p' "$GAMEDIR/version.json")
    # The extract app shows the message while torch runs; a failed rom shows an error until a
    # button is pressed, then the next is tried
    for rom in "${ROMS[@]}"; do
        r=$(printf %q "$rom")
        $LOVE_RUN "$GAMEDIR/extract" --text "Converting $(basename "$rom"), this can take a few minutes..." \
            --run "assets/torch.${DEVICE_ARCH} o2r $r ${VERSION:+-u $VERSION}"
        [ -f "sm64.o2r" ] && break
    done
    if [ ! -f "sm64.o2r" ]; then
        pm_message "No sm64.o2r created! A Super Mario 64 US or JP rom is required."
        sleep 5
        pm_finish
        exit 0
    fi
}

# Temporary swap for the rom conversion on <=1GB devices (swap_on/swap_off)
source "$GAMEDIR/swap.txt"

# --------------------- END FUNCTIONS ---------------------

# The settings launcher edits ghostship.cfg.json line by line, so seed a pretty-printed one
# (controller navigation for the in-game menu)
[ -f ghostship.cfg.json ] || printf '{\n    "CVars": {\n        "gSettings": {\n            "ControlNav": 1\n        }\n    }\n}\n' > ghostship.cfg.json

# Launcher (Love2D): frame rate and N64 low-res mode; also closes the menu bar and enables
# controller navigation in ghostship.cfg.json
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/ghostship.cfg.json" "$GAMEDIR/launch.cfg"

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
extract_roms
swap_off

# Run the game
$GPTOKEYB "Ghostship.${DEVICE_ARCH}" -c "ghostship.gptk" &
pm_platform_helper "Ghostship.${DEVICE_ARCH}" >/dev/null
./Ghostship.${DEVICE_ARCH}

# Cleanup
swap_off
rm -rf "$GAMEDIR/logs/"
pm_finish
