#!/bin/bash
# check_port.sh
#
# Static checks on a port folder (new_ports/<port>/<port>/) for the problems
# PortMaster-New reviewers reject ports for. Errors fail (exit 1); everything
# else is printed as a ::warning::. package.sh runs it before zipping, so CI
# runs it on every build.
#
# Usage (from the repo root):
#   recipes/templates/scripts/check_port.sh <port>

set -uo pipefail

port="${1:?Usage: $0 <port>}"
recipe="recipes/ports/$port/recipe.json"
root="new_ports/$port/$port"        # what goes in <port>.zip
dir="$root/$port"                   # the port's data folder
errors=0

err() { echo "::error::$port: $*"; errors=$((errors + 1)); }
warn() { echo "::warning::$port: $*"; }

[[ -f "$recipe" ]] || { err "no $recipe"; exit 1; }
pj="$dir/port.json"
[[ "$(jq -r .port_json "$recipe")" == "$pj" ]] || err "recipe port_json should be $pj"
jq -e . "$pj" > /dev/null 2>&1 || { err "$pj is missing or not valid JSON"; exit 1; }

# ---------------------------------------------------------------- port.json and the zip root
[[ "$(jq -r .name "$pj")" == "$port.zip" ]] || err "port.json name should be $port.zip"
launcher=$(jq -r '.items[0] // ""' "$pj")
[[ "$launcher" == *.sh ]] || err "port.json items[0] should be the launcher .sh"
[[ "$(jq -r '.items[1] // "" | rtrimstr("/")' "$pj")" == "$port" ]] || err "port.json items[1] should be \"$port\""
while IFS= read -r item; do
    [[ -e "$root/$item" ]] || err "port.json lists \"$item\" but it is not in $root"
done < <(jq -r '.items[]' "$pj")
while IFS= read -r f; do
    jq -e --arg f "$f" '[.items[], (.items_opt // [])[]] | map(rtrimstr("/")) | index($f)' "$pj" > /dev/null \
        || err "\"$f\" is in $root but not in port.json items"
done < <(find "$root" -mindepth 1 -maxdepth 1 ! -name .gitkeep -printf '%f\n')
[[ -d "$dir/$port" ]] && err "nested duplicate folder $dir/$port"

# ---------------------------------------------------------------- launcher and .gptk
# the launcher and anything it sources or hands to gptokeyb
while IFS= read -r f; do
    [[ -f "$f" ]] && grep -q $'\r' "$f" && err "CRLF line endings in ${f#"$root"/} (black screen, no log)"
done < <(printf '%s\n' "$root"/*.sh "$dir"/*.sh "$dir"/*.gptk; jq -r --arg d "$dir" '.includes[]? | select(endswith(".txt")) | "\($d)/\(.)"' "$recipe")

sh="$root/$launcher"
if [[ -f "$sh" ]]; then
    code=$(grep -vE '^[[:space:]]*#' "$sh")
    grep -qE '(^|[;&|({[:space:]])sudo[[:space:]]' <<< "$code" && err "$launcher uses sudo: use \$ESUDO"
    grep -qE '(pkill|killall)[^#]*gptokeyb' <<< "$code" && err "$launcher kills gptokeyb itself: pm_finish does that"
    grep -qE '^[[:space:]]*trap[[:space:]]' <<< "$code" && warn "$launcher sets a trap: make sure pm_finish still runs"
    grep -qE '(^|[[:space:]])pm_finish([[:space:]]|$)' <<< "$code" || err "$launcher does not call pm_finish"
    [[ "$(grep -vE '^[[:space:]]*$' <<< "$code" | tail -n 1 | xargs)" == pm_finish ]] \
        || warn "$launcher does not end with pm_finish"
    grep -q 'get_controls' <<< "$code" || warn "$launcher does not call get_controls"
    [[ "$(git ls-files -s -- "$sh" | cut -c1-6)" == 100755 ]] && warn "$launcher is committed as 755 (PortMaster uses 644)"
    if command -v shellcheck > /dev/null; then
        while IFS= read -r l; do err "shellcheck $l"; done < <(shellcheck -s bash -S error -f gcc "$sh")
        # SC1090/1091 sourced files, SC2154 variables from control.txt, SC2164 cd without || exit
        shellcheck -s bash -S warning -e SC1090,SC1091,SC2154,SC2164,SC2034 -f gcc "$sh" \
            | while IFS= read -r l; do warn "shellcheck $l"; done   # warnings: a subshell is fine
    else
        echo "check_port: shellcheck not installed, launcher lint skipped"
    fi
else
    err "launcher $launcher not found"
fi

# ---------------------------------------------------------------- bundled libraries
# CFWs ship their own SDL2, C/C++ runtime and GPU stack; bundled copies break on other devices
forbidden='^(libSDL2(-2\.0)?\.so|libc\.so|libm\.so|libdl\.so|libpthread\.so|librt\.so|libstdc\+\+\.so|libgcc_s\.so|ld-linux|libGL(X|ESv1_CM|ESv2|dispatch)?\.so|libEGL\.so|libgbm\.so|libdrm\.so|libglapi\.so|libmali|libMali|libvulkan\.so|libwayland-(client|server|egl)\.so)'
while IFS= read -r -d '' f; do
    [[ "$(basename "$f")" =~ $forbidden ]] && err "bundles ${f#"$dir"/}: the CFW provides it"
done < <(find "$dir" -path "$dir/libs.*" \( -type f -o -type l \) -print0)

is_elf() { [[ -f "$1" && "$(head -c 4 "$1" | od -An -c | tr -d ' ')" == '177ELF' ]]; }
for f in "$dir"/*.aarch64 "$dir"/*.x86_64; do
    is_elf "$f" || continue
    readelf -W --dyn-syms "$f" 2> /dev/null | awk '$7 != "UND" && $8 == "SDL_Init"' | grep -q . \
        && err "${f#"$dir"/} has SDL2 linked in statically: link against the system libSDL2"
done

# ---------------------------------------------------------------- glibc
# The aarch64 builder is Ubuntu 20.04 (glibc 2.31); a blank min_glibc means "runs anywhere"
baseline=2.31
need=$( { for f in "$dir"/*.aarch64; do is_elf "$f" && echo "$f"; done
          find "$dir/libs.aarch64" -type f 2> /dev/null; } | while IFS= read -r f; do
            is_elf "$f" && readelf -W --version-info "$f" 2> /dev/null; done \
        | grep -o 'GLIBC_[0-9][0-9.]*' | sed 's/GLIBC_//' | sort -uV | tail -n 1)
min=$(jq -r '.attr.min_glibc // ""' "$pj")
vle() { [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n 1)" == "$1" ]]; }
if [[ -z "$need" ]]; then
    jq -e '.attr.arch | index("aarch64")' "$pj" > /dev/null && warn "no aarch64 binaries in $dir (not built yet?)"
elif [[ -z "$min" ]]; then
    vle "$need" "$baseline" || err "aarch64 binaries need glibc $need: set attr.min_glibc to \"$need\" or build on Ubuntu 20.04"
elif ! vle "$need" "$min"; then
    err "aarch64 binaries need glibc $need but attr.min_glibc is $min"
elif [[ "$need" != "$min" ]]; then
    warn "attr.min_glibc is $min but the binaries only need $need (devices in between are excluded)"
fi

# ---------------------------------------------------------------- runtimes
# love_11.5 ships inside PortMaster itself (PortMaster/runtimes/), so it is not a runtimes.json key
builtin_runtimes='love_11.5'
mapfile -t runtimes < <(jq -r '.attr.runtime // [] | if type == "array" then .[] else "NOT_AN_ARRAY" end' "$pj")
if (( ${#runtimes[@]} )); then
    rj=$(curl -sfL --max-time 20 https://raw.githubusercontent.com/PortsMaster/PortMaster-New/main/runtimes/runtimes.json)
    for r in "${runtimes[@]}"; do
        if [[ "$r" == NOT_AN_ARRAY ]]; then err "attr.runtime must be a JSON array"
        elif [[ " $builtin_runtimes " == *" $r "* ]]; then warn "attr.runtime lists $r, which ships with PortMaster and is not in runtimes.json: remove it"
        elif [[ -z "$rj" ]]; then warn "could not download runtimes.json, attr.runtime $r not checked"
        elif ! jq -e --arg r "$r" 'has($r)' <<< "$rj" > /dev/null; then err "attr.runtime $r is not a key in PortMaster's runtimes.json"
        fi
    done
fi

# ---------------------------------------------------------------- gameinfo.xml
gi="$dir/gameinfo.xml"
if [[ -f "$gi" ]]; then
    while IFS= read -r l; do err "$l"; done < <(python3 -I - "$gi" "$root" "./$launcher" <<'EOF'
import os, sys, xml.etree.ElementTree as ET
gi, root, launcher = sys.argv[1:]
try:
    games = ET.parse(gi).getroot().findall("game")
except ET.ParseError as e:
    sys.exit(print("gameinfo.xml is not valid XML: %s" % e))
for g in games:
    if (g.findtext("path") or "") != launcher:
        print("gameinfo.xml <path> is %r, the launcher is %r" % (g.findtext("path"), launcher))
    for tag in ("image", "video", "thumbnail", "marquee"):
        v = g.findtext(tag)
        if v and not os.path.exists(os.path.join(root, v)):
            print("gameinfo.xml <%s> %s does not exist" % (tag, v))
if not games:
    print("gameinfo.xml has no <game>")
EOF
)
else
    err "no gameinfo.xml"
fi

# ---------------------------------------------------------------- licenses
if [[ -d "$dir/licenses" ]]; then
    find "$dir/licenses" -mindepth 1 -type d | grep -q . && err "licenses/ has subfolders: keep it flat"
    odd=$(find "$dir/licenses" -maxdepth 1 -type f ! -name 'LICENSE.*.txt' ! -name .gitkeep -printf '%f ')
    [[ -n "$odd" ]] && warn "licenses/ names should be LICENSE.<component>.txt: $odd"
    find "$dir/licenses" -maxdepth 1 -type f ! -name .gitkeep | grep -q . || err "licenses/ is empty"
else
    err "no licenses/ folder"
fi

# ---------------------------------------------------------------- screenshot and cover
png_size() { od -An -tu1 -j16 -N8 "$1" | awk '{printf "%dx%d", $1*16777216+$2*65536+$3*256+$4, $5*16777216+$6*65536+$7*256+$8}'; }
for img in screenshot.png cover.png; do
    f="$dir/$img"
    if [[ ! -f "$f" ]]; then warn "no $img (make_images.sh creates a placeholder)"; continue; fi
    [[ "$img" == screenshot.png && "$(png_size "$f")" != 640x480 ]] && warn "$img is $(png_size "$f"), not 640x480"
    if grep -qa 'PortMaster placeholder' "$f"; then
        if [[ "$img" == screenshot.png ]] && jq -e .screenshot "$recipe" > /dev/null; then
            warn "$img is still the placeholder although recipe.json has a screenshot config: the Xvfb capture failed (see Capture screenshot)"
        else
            warn "$img is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)"
        fi
    fi
done

if (( errors )); then
    echo "check_port: $port: $errors error(s)"
    exit 1
fi
echo "check_port: $port: ok"
