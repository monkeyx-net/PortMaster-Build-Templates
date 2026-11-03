#!/bin/bash
# package.sh
#
# Turns new_ports/<port>/<port>/ into <port>.zip, the same way in CI and locally:
# copies the recipe's shared includes into the port folder, writes version.json
# from the recipe, runs check_port.sh (errors stop here) and zips the folder
# without .gitkeep files. The binaries must already be in the port folder
# (CI extracts them from the build artifacts first).
#
# Usage (from the repo root):
#   recipes/templates/scripts/package.sh <port> [out.zip]     default: ./<port>.zip

set -euo pipefail

port="${1:?Usage: $0 <port> [out.zip]}"
out="$(realpath -m "${2:-$port.zip}")"
recipe="recipes/ports/$port/recipe.json"
dir="$(dirname "$(jq -r .port_json "$recipe")")"
scripts="$(dirname "$0")"

for inc in $(jq -r '.includes[]?' "$recipe"); do
    mkdir -p "$dir/$(dirname "$inc")"
    cp "recipes/templates/includes/$inc" "$dir/$inc"
done

jq '{name, version: .source.port_version}' "$recipe" > "$dir/version.json"

"$scripts/check_port.sh" "$port"

rm -f "$out"
cd "new_ports/$port/$port"
if command -v zip > /dev/null; then
    zip -qr "$out" . -x '*.gitkeep'
else
    # same entries as zip -r (directories included), for machines without zip
    python3 -I - "$out" <<'EOF'
import os, sys, zipfile
with zipfile.ZipFile(sys.argv[1], "w", zipfile.ZIP_DEFLATED) as z:
    for top, dirs, files in os.walk("."):
        dirs.sort()
        for d in dirs:
            z.write(os.path.join(top, d)[2:], os.path.join(top, d)[2:] + "/")
        for f in sorted(files):
            if f != ".gitkeep":
                z.write(os.path.join(top, f), os.path.join(top, f)[2:])
EOF
fi
echo "Packaged $out"
