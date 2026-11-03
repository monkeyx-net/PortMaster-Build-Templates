#!/bin/bash
# test_port.sh
#
# Runs a port on real handhelds through DeviceHub (the devhub command) and
# writes new_ports/<port>/test/report.md: one row per device with CFW, SoC/GPU,
# resolution, result, FPS and a screenshot, for the PortMaster-New PR.
#
# Per device: take the lock, install the zip with the device's PortMaster,
# stop the front end, start the launcher the way the front end does, wait,
# press the recipe's test keys, take a screenshot, check the binaries with ldd,
# pull the logs, then stop the game, restart the front end and give the lock back.
# PASS = launcher still running at the screenshot, screenshot not blank, no
# missing libraries.
#
# Optional "test" object in recipe.json (same fields as "screenshot"):
#   "test": {
#     "delay": 15,                   seconds before the keys/screenshot (default 15)
#     "keys": "a 3 start",           devhub input tokens: a b x y l1 r1 start select
#                                    up down left right key:enter ...; a number waits
#                                    (a b x y are positional: Nintendo-layout devices may
#                                    swap them, so prefer start, the D-pad and key:)
#     "env": "FOO=1",                extra environment for the launcher
#     "needs_data": "oot.o2r baseroms/*.z64"   user-supplied files (globs in the port
#                                    folder); without them only the launcher's error
#                                    path is tested. true = can't be checked.
#   }
#
# Usage (from the repo root):
#   recipes/templates/scripts/test_port.sh <port> [--zip FILE|URL] [--screenshot DEV] [DEV...]
#     no DEV: every online device that is not out of service
#     --zip: test this zip/URL instead of packaging new_ports/<port>/<port>/ with package.sh
#     --screenshot DEV: use DEV's screenshot (resized to 640x480) as the port's screenshot.png
#       and drop the recipe's Xvfb "screenshot" config; pick a 4:3 640x480 device
# DEVHUB=/path/to/devhub if devhub is not on PATH.

set -uo pipefail

port="${1:?Usage: $0 <port> [--zip FILE|URL] [--screenshot DEV] [DEV...]}"
shift
zip="" shotdev="" devs=()
while (( $# )); do
    case "$1" in
        --zip) zip="$2"; shift 2 ;;
        --screenshot) shotdev="$2"; shift 2 ;;
        *) devs+=("$1"); shift ;;
    esac
done

DH="${DEVHUB:-devhub}"
command -v "$DH" > /dev/null || { echo "devhub not found: put it on PATH or set DEVHUB=/path/to/DeviceHub/devhub"; exit 2; }
export DEVHUB_NAME="${DEVHUB_NAME:-test_port.sh $port}"

recipe="recipes/ports/$port/recipe.json"
pj="$(jq -r .port_json "$recipe")"
dir="$(dirname "$pj")"
launcher="$(jq -r '.items[0]' "$pj")"
delay=$(jq -r '.test.delay // 15' "$recipe")
keys=$(jq -r '.test.keys // ""' "$recipe")
env=$(jq -r '.test.env // ""' "$recipe")
needs_data=$(jq -r '.test.needs_data // false | if . == true then "?" elif . == false then "" else . end' "$recipe")
out="new_ports/$port/test"
rm -rf "$out"      # results of this run only
mkdir -p "$out"
built=""
trap '[[ -n "$built" ]] && rm -f "$built"' EXIT     # the zip we packaged: installed already, results stay

if [[ -z "$zip" ]]; then
    zip="$out/$port.zip" built="$zip"
    recipes/templates/scripts/package.sh "$port" "$zip" || { echo "package.sh failed: fix the errors above first"; exit 1; }
fi

status=$("$DH" status --json) || { echo "DeviceHub is not reachable"; exit 3; }
if (( ! ${#devs[@]} )); then
    mapfile -t devs < <(jq -r '.[] | select(.rt.online == true and (.oos | not) and .enabled) | .id' <<< "$status")
fi
[[ -n "$shotdev" && " ${devs[*]} " != *" $shotdev "* ]] && devs+=("$shotdev")
(( ${#devs[@]} )) || { echo "no online devices"; exit 1; }
info() { jq -r --arg d "$1" ".[] | select(.id == \$d) | $2 // \"\"" <<< "$status"; }

# pgrep pattern for the launcher that does not match the shell running pgrep
pat=$(sed 's/[][\.*^$()+?{}|]/\\&/g' <<< "$launcher")
pat="[${pat:0:1}]${pat:1}"

# Ask for every lock at once (each may wait for the user's approval), then test one device at a time
echo "Requesting locks on: ${devs[*]}"
declare -A lockrc
for dev in "${devs[@]}"; do
    "$DH" acquire "$dev" "test $port" --for 30m --wait 20m --need "multi-device test of $port" > "$out/$dev.lock.txt" 2>&1 &
    lockrc[$dev]=$!
done
for dev in "${devs[@]}"; do
    wait "${lockrc[$dev]}"; lockrc[$dev]=$?
    cat "$out/$dev.lock.txt"
done

rows=()
write_report() {
    {
        echo "## $(jq -r .attr.title "$pj") on devices"
        echo
        echo "Tested $(date +%Y-%m-%d) with $(basename "$zip") (version $(jq -r .source.port_version "$recipe"))."
        echo
        echo "| Device | CFW | SoC / GPU | Resolution | Result | Avg FPS | 1% low | Screenshot |"
        echo "|---|---|---|---|---|---|---|---|"
        printf '%s\n' "${rows[@]}"
    } > "$out/report.md"
}

cur="" cur_sid="" cur_input=0
cleanup() {      # leave the device as we found it: game stopped, front end up, lock released
    local dev="$cur"
    [[ -z "$dev" ]] && return
    (( cur_input )) && "$DH" input "$dev" close > /dev/null
    "$DH" fps "$dev" stop > /dev/null 2>&1
    if [[ -n "$cur_sid" ]]; then
        "$DH" run "$dev" -- "S=''; [ \"\$(id -u)\" = 0 ] || S='sudo -n'; \$S pkill -TERM -s $cur_sid; sleep 2; \$S pkill -KILL -s $cur_sid; true" > /dev/null
    fi
    "$DH" fe "$dev" restart > /dev/null
    "$DH" release "$dev" || "$DH" release "$dev" --dirty "test_port.sh: front end did not come back after the test"
    cur="" cur_sid="" cur_input=0
}
trap 'echo "interrupted"; cleanup; write_report; exit 130' INT TERM

for dev in "${devs[@]}"; do
    row_info="| $(info "$dev" .type) ($dev) | $(info "$dev" .cfw) | $(info "$dev" .soc) / $(info "$dev" .gpu) | $(info "$dev" .screen) "
    if [[ "${lockrc[$dev]}" != 0 ]]; then
        rows+=("$row_info| not tested (no lock) | | | |"); write_report; continue
    fi
    cur="$dev"
    echo "=== $(date +%T) $dev: installing $(basename "$zip")"
    if ! "$DH" install "$dev" "$zip"; then
        rows+=("$row_info| FAIL: install | | | |"); cleanup; write_report; continue
    fi
    ports_dir=$("$DH" ports "$dev" --json | jq -r '.ports_dir // ""')
    [[ -n "$ports_dir" ]] || ports_dir=$(info "$dev" .ports_dir)
    gamedir="$ports_dir/$port"

    nodata=0
    if [[ "$needs_data" == "?" ]]; then
        nodata=1
    elif [[ -n "$needs_data" ]]; then
        "$DH" run "$dev" --ro -- "cd '$gamedir' && ls -d $needs_data" > /dev/null 2>&1 || nodata=1
    fi

    "$DH" fe "$dev" stop
    if [[ -n "$keys" ]]; then      # before the launch: some games only look for controllers at start
        "$DH" input "$dev" open && cur_input=1
    fi
    "$DH" fps "$dev" start "$port test" > /dev/null
    echo "=== $(date +%T) $dev: starting $launcher, screenshot in ${delay}s"
    # cd separately: "cd && cmd &" backgrounds a subshell that keeps devhub run waiting for the game
    cur_sid=$("$DH" run "$dev" -- "cd '$ports_dir' || exit 1; setsid nohup env $env bash './$launcher' > /tmp/test_$port.log 2>&1 < /dev/null & echo \$!" | tail -n 1)
    sleep "$delay"
    if [[ -n "$keys" ]]; then
        "$DH" input "$dev" $keys
        sleep 2
    fi

    echo "=== $(date +%T) $dev: checking the game, screenshot"
    reasons=()
    alive=$("$DH" run "$dev" --ro -- "pgrep -f '$pat' > /dev/null && echo alive")
    [[ "$alive" == *alive* ]] || reasons+=("launcher exited")
    rm -f "$out/$dev.png"
    "$DH" shot "$dev" "$out/$dev.png" || reasons+=("no screenshot")
    if [[ -f "$out/$dev.png" ]] && [[ $(convert "$out/$dev.png" -format '%[fx:standard_deviation < 0.02]' info:) == 1 ]]; then
        reasons+=("blank screen")
    fi

    # same library path as the launchers: libs.<arch> first, libs.<arch>/fallback last
    "$DH" run "$dev" --ro -- "cd '$gamedir' && a=\$(uname -m) && L=ldd && { command -v ldd > /dev/null || L=\"\$(ls /lib/ld-linux-*.so.* /lib64/ld-linux-*.so.* 2> /dev/null | head -n 1) --list\"; }
        for b in *.\$a; do [ -f \"\$b\" ] || continue; echo \"== \$b\"; LD_LIBRARY_PATH=\"\$PWD/libs.\$a:\$LD_LIBRARY_PATH:\$PWD/libs.\$a/fallback\" \$L \"./\$b\" 2>&1; done" \
        > "$out/$dev.ldd.txt" 2>&1
    missing=$(grep 'not found' "$out/$dev.ldd.txt" | awk '{print $1}' | sort -u | xargs)
    [[ -n "$missing" ]] && reasons+=("missing libs: $missing")

    "$DH" pull "$dev" "$gamedir/log.txt" "$out/$dev.log.txt" > /dev/null 2>&1
    "$DH" pull "$dev" "/tmp/test_$port.log" "$out/$dev.launch.txt" > /dev/null 2>&1
    fps=$("$DH" fps "$dev" stop 2> /dev/null)
    avg=$(grep -oP 'avg +\K[0-9.]+' <<< "$fps")
    low=$(grep -oP '1% low +\K[0-9.]+' <<< "$fps")
    cleanup

    result="PASS"
    if (( nodata )); then
        # no game data: the launcher shows its message/settings screen or exits; only the screen and libs count
        keep=()
        for r in "${reasons[@]}"; do [[ "$r" == "launcher exited" ]] || keep+=("$r"); done
        reasons=("${keep[@]}")
        result="launcher only (no game data)"
    fi
    if (( ${#reasons[@]} )); then
        result="FAIL: $(printf '%s; ' "${reasons[@]}")"
        result="${result%; }"
    fi
    echo "=== $(date +%T) $dev: $result (avg ${avg:-?} fps, 1% low ${low:-?})"
    rows+=("$row_info| $result | ${avg} | ${low} | [$dev.png]($dev.png) |")
    write_report

    if [[ "$dev" == "$shotdev" ]]; then
        if [[ "$result" == FAIL* ]]; then
            echo "Not using $dev's screenshot: the test failed"
        else
            convert "$out/$dev.png" -resize 640x480 -background black -gravity center -extent 640x480 \
                -alpha off -depth 8 "$dir/screenshot.png"
            grep -qa 'PortMaster placeholder' "$dir/cover.png" 2> /dev/null && cp "$dir/screenshot.png" "$dir/cover.png"
            if jq -e .screenshot "$recipe" > /dev/null; then
                jq 'del(.screenshot)' "$recipe" > "$recipe.tmp" && mv "$recipe.tmp" "$recipe"
                echo "Removed the Xvfb screenshot config from $recipe (CI would overwrite the device screenshot)"
            fi
            [[ "$(info "$dev" .screen)" == 640x480 ]] || echo "Note: $dev is $(info "$dev" .screen), not 640x480: the screenshot is letterboxed"
            echo "Wrote $dir/screenshot.png from $dev"
        fi
    fi
done
trap - INT TERM
echo
cat "$out/report.md"
grep -q '| FAIL' "$out/report.md" && exit 1
exit 0
