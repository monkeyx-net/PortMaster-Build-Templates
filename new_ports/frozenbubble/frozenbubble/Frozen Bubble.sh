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

GAMEBINARY=frozen-bubble-sdl2
GAMEDIR=/$directory/ports/frozenbubble
> "$GAMEDIR/log.txt" && exec > >(tee "$GAMEDIR/log.txt") 2>&1

cd $GAMEDIR

# Game settings (settings.ini, high scores) live in conf/
bind_directories ~/.local/share/frozen-bubble $GAMEDIR/conf

# IP address shown in the launcher when hosting a server
DEVICE_IP=$(ip -4 route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<NF;i++) if($i=="src") print $(i+1)}' | head -n1)
[ -z "$DEVICE_IP" ] && DEVICE_IP=$(hostname -I 2>/dev/null | awk '{print $1}')

# Network / multiplayer launcher (Love2D)
rm -f "$GAMEDIR/launch.cfg"
source $controlfolder/runtimes/"love_11.5"/love.txt
$LOVE_RUN "$GAMEDIR/launcher" "$GAMEDIR/conf" "$GAMEDIR/launch.cfg" "$DEVICE_IP"

FB_LAUNCH=0
[ -f "$GAMEDIR/launch.cfg" ] && source "$GAMEDIR/launch.cfg"
if [ "$FB_LAUNCH" != "1" ]; then
  echo "Launcher exited without starting the game"
  pm_finish
  exit 0
fi

# libs.<arch>/fallback (glib for fb-server) is searched after the system libraries
export LD_LIBRARY_PATH="$GAMEDIR/libs.${DEVICE_ARCH}:$LD_LIBRARY_PATH:$GAMEDIR/libs.${DEVICE_ARCH}/fallback"
export FB_DATA_DIR="$GAMEDIR/share"
export FB_SERVER_BIN="$GAMEDIR/fb-server.${DEVICE_ARCH}"
chmod +x "$GAMEDIR/$GAMEBINARY.${DEVICE_ARCH}" "$FB_SERVER_BIN" 2>/dev/null

SERVER_PID=""
if [ "$FB_HOST_SERVER" == "1" ]; then
  pkill -f "fb-server.${DEVICE_ARCH}" 2>/dev/null
  SERVER_ARGS=(-q -z -p "$FB_SERVER_PORT" -m "$FB_SERVER_MAX")
  [ -n "$FB_SERVER_NAME" ] && SERVER_ARGS+=(-n "$FB_SERVER_NAME")
  [ "$FB_SERVER_LAN" == "1" ] && SERVER_ARGS+=(-l)
  echo "Starting server: $FB_SERVER_BIN ${SERVER_ARGS[*]}"
  # -d keeps it in the foreground so it can be stopped when the game exits
  "$FB_SERVER_BIN" -d "${SERVER_ARGS[@]}" > "$GAMEDIR/server.log" 2>&1 &
  SERVER_PID=$!
  sleep 1
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "Server failed to start, see server.log"
    SERVER_PID=""
  fi
fi

# The game reads controllers natively; gptokeyb only provides the exit hotkey
$GPTOKEYB "$GAMEBINARY.${DEVICE_ARCH}" &
SDL_GAMECONTROLLERCONFIG="$sdl_controllerconfig" ./$GAMEBINARY.${DEVICE_ARCH}

if [ -n "$SERVER_PID" ]; then
  kill "$SERVER_PID" 2>/dev/null
fi
pkill -f "fb-server.${DEVICE_ARCH}" 2>/dev/null
pm_finish
