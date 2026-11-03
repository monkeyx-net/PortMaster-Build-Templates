#!/bin/bash
# make_images.sh
#
# Creates screenshot.png and cover.png (640x480) in a port folder if they are
# missing. If only one exists it is copied to the other; if both are missing a
# title card is generated. Generated images are placeholders: replace them with
# a real screenshot before submitting to PortMaster. They carry the PNG comment
# "PortMaster placeholder", which check_port.sh looks for.
#
# Usage (from the repo root):
#   recipes/templates/scripts/make_images.sh <port data folder> "<Port Title>"
#   e.g. make_images.sh new_ports/recwars/recwars/recwars "Rec Wars"

set -euo pipefail

dir="${1:?Usage: $0 <port data folder> <title>}"
title="${2:?Usage: $0 <port data folder> <title>}"
shot="$dir/screenshot.png"
cover="$dir/cover.png"

if [[ -f "$shot" && -f "$cover" ]]; then
    exit 0
elif [[ -f "$shot" ]]; then
    cp "$shot" "$cover"; echo "Created $cover from screenshot.png"; exit 0
elif [[ -f "$cover" ]]; then
    cp "$cover" "$shot"; echo "Created $shot from cover.png"; exit 0
fi

im=$(command -v magick || command -v convert) || { echo "ImageMagick is required"; exit 1; }
font=$(fc-match -f '%{file}' 'DejaVu Sans:bold' 2>/dev/null || true)

"$im" -size 640x480 gradient:'#1e2a3a-#0b0f16' \
    \( -size 560x320 -background none -fill white ${font:+-font "$font"} \
       -gravity center caption:"$title" \) \
    -gravity center -composite \
    -fill '#8a9bb0' ${font:+-font "$font"} -pointsize 20 -gravity south \
    -annotate +0+24 'PortMaster' \
    -set comment 'PortMaster placeholder' -alpha off -depth 8 "$shot"
cp "$shot" "$cover"
echo "::warning::$title: generated placeholder screenshot.png and cover.png -- replace with a real 640x480 screenshot or add a screenshot config to recipe.json"
