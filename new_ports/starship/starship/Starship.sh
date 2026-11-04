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

# Temporary swap for rom extraction on <=1GB devices (swap_on/swap_off)
source "$GAMEDIR/swap.txt"

# --------------------- END FUNCTIONS ---------------------

# Launcher (Love2D): frame rate, low-res mode, voices, invert Y; also closes the menu bar and
# enables controller navigation in starship.cfg.json
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/starship.cfg.json" "$GAMEDIR/launch.cfg"

LAUNCHER_PLAY=0
[ -f "$GAMEDIR/launch.cfg" ] && source "$GAMEDIR/launch.cfg"
if [ "$LAUNCHER_PLAY" != "1" ]; then
  echo "Launcher exited without starting the game"
  pm_finish
  exit 0
fi

# Perform functions
unzip_assets
find_roms ||{ pm_finish; exit 0; }
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
