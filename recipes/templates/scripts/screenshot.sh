#!/bin/bash
# screenshot.sh
#
# Runs a port's x86_64 build under Xvfb and captures a 640x480 screenshot.png.
# Opt-in per port via a "screenshot" object in recipe.json:
#
#   "screenshot": {
#     "delay": 8,                      seconds to wait after starting (default 8)
#     "keys": "Return Down Return",    xdotool keys pressed 1s apart, then 2s wait
#     "args": "--fullscreen",          extra arguments for the game
#     "env": "FOO=1 BAR=2"             extra environment variables
#   }
#
# Runs inside the x86_64 builder container as root (installs xvfb/xdotool if
# missing). Expects /workspace/<port>-linux-x86_64.tar.gz from build.sh and
# writes /workspace/<port>-linux-screenshot.tar.gz, which the package job
# extracts into the port folder like the binaries. A blank capture is skipped.
#
# Usage (from the repo root):
#   recipes/templates/scripts/screenshot.sh <port>

set -euo pipefail

port="${1:?Usage: $0 <port>}"
recipe="recipes/ports/$port/recipe.json"

jq -e '.screenshot' "$recipe" > /dev/null || { echo "No screenshot config in $recipe -- skipping"; exit 0; }
delay=$(jq -r '.screenshot.delay // 8' "$recipe")
keys=$(jq -r '.screenshot.keys // ""' "$recipe")
args=$(jq -r '.screenshot.args // ""' "$recipe")
env=$(jq -r '.screenshot.env // ""' "$recipe")
exe="$(jq -r '.source.port_exe' "$recipe").x86_64"
port_dir=$(dirname "$(jq -r '.port_json' "$recipe")")

if ! command -v Xvfb > /dev/null || ! command -v xdotool > /dev/null; then
    apt-get update -qq && apt-get install -y -qq --no-install-recommends xvfb xdotool > /dev/null
fi

game_pid=""
work=$(mktemp -d)
cp -a "$port_dir"/. "$work"/
tar -xzf "/workspace/${port}-linux-x86_64.tar.gz" -C "$work"
mkdir -p "$work/.home"

Xvfb :99 -screen 0 640x480x24 -nolisten tcp 2> /dev/null &
xvfb_pid=$!
trap 'kill $game_pid $xvfb_pid 2> /dev/null || true; rm -rf "$work"' EXIT
export DISPLAY=:99
sleep 2

cd "$work"
chmod +x "$exe"
# shellcheck disable=SC2086
env HOME="$work/.home" XDG_DATA_HOME="$work/.home" SDL_AUDIODRIVER=dummy \
    LD_LIBRARY_PATH="$work/libs.x86_64" $env "./$exe" $args > "$work/.home/game.log" 2>&1 &
game_pid=$!
sleep "$delay"

if ! kill -0 "$game_pid" 2> /dev/null; then
    echo "::warning::$port exited before the screenshot was taken"; tail -20 "$work/.home/game.log"; exit 0
fi

if [[ -n "$keys" ]]; then
    wid=$(xdotool search --onlyvisible --pid "$game_pid" 2> /dev/null | head -n 1 || true)
    [[ -n "$wid" ]] && xdotool windowfocus "$wid" || true
    for key in $keys; do xdotool key "$key"; sleep 1; done
    sleep 2
fi

out="/workspace/screenshot"
mkdir -p "$out"
import -window root "$work/.home/raw.png"
if [[ $(convert "$work/.home/raw.png" -format '%[fx:standard_deviation < 0.02]' info:) == 1 ]]; then
    echo "::warning::$port screenshot is blank -- skipping"; exit 0
fi
convert "$work/.home/raw.png" -resize 640x480 -background black -gravity center -extent 640x480 \
    -alpha off -depth 8 "$out/screenshot.png"
tar -czf "/workspace/${port}-linux-screenshot.tar.gz" -C "$out" screenshot.png
echo "Captured $port screenshot"
