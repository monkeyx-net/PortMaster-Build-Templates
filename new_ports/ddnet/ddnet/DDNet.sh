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

GAMEBINARY=DDNet
GAMEDIR=/$directory/ports/ddnet
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

cd $GAMEDIR
chmod +x "$GAMEBINARY.${DEVICE_ARCH}" "DDNet-Server.${DEVICE_ARCH}"

# Settings, maps, demos and skins (DDNet uses $XDG_DATA_HOME/ddnet)
bind_directories "$XDG_DATA_HOME/ddnet" "$GAMEDIR/conf"

# curl is linked statically without a CA bundle path: use the CFW's certificates for HTTPS
for CA in /etc/ssl/certs/ca-certificates.crt /etc/ssl/cacert.pem /etc/pki/tls/certs/ca-bundle.crt; do
  [ -f "$CA" ] && export SSL_CERT_FILE="$CA" && break
done

source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN controls/controls.love controls/controls.png 30 $DISPLAY_WIDTH $DISPLAY_HEIGHT

$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" -c "./$GAMEBINARY.gptk" &
pm_platform_helper "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}"
SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig" ./$GAMEBINARY.${DEVICE_ARCH}
pm_finish
