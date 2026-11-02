#!/bin/bash
# fetch_source.sh
#
# Downloads the upstream source recorded in recipes/ports/<port>/recipe.json
# (source.port_download), verifies it against source.port_checksum and
# extracts it to new_ports/<port>/source. Sources are not committed to git;
# the build workflow runs this before building.
#
# Usage (from the repo root):
#   recipes/templates/scripts/fetch_source.sh <port>

set -euo pipefail

port="${1:?Usage: $0 <port>}"
recipe="recipes/ports/$port/recipe.json"
dest="new_ports/$port/source"

[[ -f "$recipe" ]] || { echo "Recipe not found: $recipe"; exit 1; }

url=$(jq -r '.source.port_download // empty' "$recipe")
checksum=$(jq -r '.source.port_checksum // empty' "$recipe")
[[ -n "$url" ]] || { echo "No source.port_download in $recipe -- run check_updates.sh $port"; exit 1; }

case "${url%%\?*}" in
    *.tar.gz|*.tgz) archive="new_ports/$port/source.tar.gz" ;;
    *.zip)          archive="new_ports/$port/source.zip" ;;
    *)              echo "Unsupported archive type: $url"; exit 1 ;;
esac

echo "Downloading $url"
mkdir -p "new_ports/$port"
curl -fsSL -o "$archive" "$url"

if [[ -n "$checksum" ]]; then
    actual=$(sha256sum "$archive" | cut -d' ' -f1)
    if [[ "$actual" != "$checksum" ]]; then
        echo "Checksum mismatch for $port: expected $checksum, got $actual"
        echo "Upstream changed since the last check -- run the Check Recipe Updates workflow first."
        exit 1
    fi
fi

rm -rf "$dest"
mkdir -p "$dest"
if [[ "$archive" == *.tar.gz ]]; then
    tar -xzf "$archive" -C "$dest" --strip-components=1
else
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    unzip -q "$archive" -d "$tmp"
    cp -a "$tmp"/*/. "$dest"/
fi
echo "Extracted $port source to $dest"
