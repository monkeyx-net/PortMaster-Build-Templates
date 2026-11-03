#!/bin/bash
# make_controls.sh
#
# Generates a port's controls image with Controls4PM (recipes/templates/Controls4PM)
# when recipe.json has "controls": "<path in the port folder>", e.g.
# "launcher/controls.png" (shown by the settings launcher's Controls item) or
# "controls/controls.png" (shown by controls.love before the game).
# Input: the | Button | Action | table in the port's README.md, its .gptk and
# port.json (title), filled into recipes/templates/PM_MAPPINGS_TEMPLATE.svg.
# Needs lua and inkscape. package.sh runs it before check_port.sh.
#
# Usage (from the repo root):
#   recipes/templates/scripts/make_controls.sh <port>

set -euo pipefail

port="${1:?Usage: $0 <port>}"
recipe="recipes/ports/$port/recipe.json"
out=$(jq -r '.controls // empty' "$recipe")
[[ -z "$out" ]] && exit 0

repo="$PWD"
dir="$repo/$(dirname "$(jq -r .port_json "$recipe")")"
readme="$dir/README.md"
fail() { echo "::error::$port: $*"; exit 1; }

grep -qiE '^\|[[:space:]]*button[[:space:]]*\|' "$readme" 2> /dev/null \
    || fail "recipe.json has \"controls\" but $readme has no | Button | Action | table"
lua=$(command -v lua5.4 || command -v lua || command -v luajit) || fail "lua is not installed"
command -v inkscape > /dev/null || fail "inkscape is not installed"

args=(create -i "$repo/recipes/templates/PM_MAPPINGS_TEMPLATE.svg" -r "$readme" -p "$dir/port.json")
gptk=$(find "$dir" -maxdepth 1 -name '*.gptk' | sort | head -n 1)
[[ -n "$gptk" ]] && args+=(-c "$gptk")

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
# converter.lua is plain Lua; it writes controls.svg and controls.png to the current directory
(cd "$tmp" && "$lua" - "${args[@]}" <<EOF
package.path = "$repo/recipes/templates/Controls4PM/src/?.lua;" .. package.path
os.exit(require("converter").run(arg) and 0 or 1)
EOF
) || fail "Controls4PM converter failed"
[[ -s "$tmp/controls.png" ]] || fail "inkscape did not write controls.png"

mkdir -p "$(dirname "$dir/$out")"
cp "$tmp/controls.png" "$dir/$out"
echo "Controls image: ${dir#"$repo"/}/$out"
