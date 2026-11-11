#!/bin/bash
# check_port.sh
#
# Static checks on a port folder (new_ports/<port>/<port>/) for the problems
# PortMaster-New reviewers reject ports for. Errors fail (exit 1); everything
# else is printed as a ::warning::. package.sh runs it before zipping, so CI
# runs it on every build.
#
# Usage (from the repo root):
#   recipes/templates/scripts/check_port.sh <port> [--fix] [--fix-bloat] [--fix-license] [--test]
#
# package.sh (so every build) runs it with --fix --fix-bloat --fix-license.
#
# Writes the result to new_ports/<port>/check.md (next to test/, so it is not in the zip).
#
# --test changes nothing in the port: it lists what --fix, --fix-bloat and --fix-license would change
# ("to fix:") and the errors and warnings left for a manual fix, on screen and in check.md.
#
# --fix repairs what can be repaired instead of reporting it (package.sh passes it, so
# builds stay correct): writes the glibc the aarch64 binaries need to attr.min_glibc in
# port.json, letterboxes screenshot.png to 640x480 and compresses screenshot/cover to <100 kB,
# and for warnings: sets the launcher to 644, fixes shellcheck SC2188 (bare "> file"),
# drops built-in runtimes from attr.runtime and renames LICENSE_x/LICENSE-x to LICENSE.x.txt.
#
# --fix-license downloads the license of each bundled lib that has none from license_urls in
# port_rules.json into licenses/LICENSE.<name>.txt (needs network; check the files).
#
# --fix-bloat deletes bundled libs that are in common_libs (the "bloats the port" warning).
# The build removes them every time, so also take them out of FILES in the port's build.sh.

set -uo pipefail

usage="Usage: $0 <port> [--fix] [--fix-bloat] [--fix-license] [--test]"
port="${1:?$usage}"
args="${*:2}"
fix="" bloat="" license=""   # fix: --fix or --test (--test wins); bloat/license: 1 with --fix-bloat/--fix-license
for a in "${@:2}"; do
    case "$a" in
        --fix) [[ "$fix" == --test ]] || fix=--fix ;;
        --test) fix=--test ;;
        --fix-bloat) bloat=1 ;;
        --fix-license) license=1 ;;
        *) echo "$usage"; exit 2 ;;
    esac
done
recipe="recipes/ports/$port/recipe.json"
root="new_ports/$port/$port"        # what goes in <port>.zip
dir="$root/$port"                   # the port's data folder
report="new_ports/$port/check.md"
errors=0 errs=() warns=() notes=()

err() { echo "::error::$port: $*"; errors=$((errors + 1)); errs+=("$*"); }
warn() { echo "::warning::$port: $*"; warns+=("$*"); }
note() { echo "check_port: $port: $*"; notes+=("$*"); }   # what --fix changed, skipped checks

# A repair: fixit FLAG MSG CMD... runs CMD (printing "fixed: MSG"), or with --test only
# collects MSG; the changes are listed at the end grouped by the flag that makes them
fixing() { [[ -n "$fix" ]]; }
declare -A changed=()   # flag -> newline-separated messages
fix_flags=(--fix --fix-bloat --fix-license)
fixit() {
    local flag="$1" msg="$2"; shift 2
    if [[ "$fix" == --test ]]; then changed[$flag]+="$msg"$'\n'; changes=$((changes + 1))
    else "$@" && echo "check_port: $port: fixed: $msg" && changed[$flag]+="$msg"$'\n' && changes=$((changes + 1)); fi
}
changes=0
# "N for --fix, M for --fix-bloat, ..." of the flags with changes
change_counts() {
    local out="" fl n
    for fl in "${fix_flags[@]}"; do
        n=$(printf '%s' "${changed[$fl]:-}" | grep -c .)
        (( n )) && out+="${out:+, }$n for $fl"
    done
    echo "${out:-none}"
}
set_mode_644() { chmod 644 "$1" && git update-index --chmod=-x -- "$1"; }
set_min_glibc() {
    if grep -q '"min_glibc":' "$pj"; then
        sed -i -E "s/(\"min_glibc\":[[:space:]]*)(\"[^\"]*\"|null)/\1\"$1\"/" "$pj"
    else
        jq --indent 2 --arg v "$1" '.attr.min_glibc = $v' "$pj" > "$pj.tmp" && mv "$pj.tmp" "$pj"
    fi
}
drop_runtime() { jq --indent 2 --arg r "$1" '.attr.runtime -= [$r]' "$pj" > "$pj.tmp" && mv "$pj.tmp" "$pj"; }

# check.md, written on every exit (also the early ones)
write_report() {
    [[ -d "new_ports/$port" ]] || return
    local result="ok" m
    (( errors )) && result="$errors error(s)"
    [[ "$fix" == --test ]] && result="dry run, changes: $(change_counts); $errors error(s) left"
    {
        echo "# check_port: $port"
        echo
        echo "\`check_port.sh $port${args:+ $args}\` on $(date '+%Y-%m-%d %H:%M'), commit $(git rev-parse --short HEAD 2> /dev/null || echo unknown)."
        echo
        echo "**Result: $result, ${#warns[@]} warning(s)**"
        local fl title lines=()
        for section in Errors Warnings "${fix_flags[@]}" Notes; do
            title="$section"
            case $section in
                Errors) lines=("${errs[@]}") ;; Warnings) lines=("${warns[@]}") ;; Notes) lines=("${notes[@]}") ;;
                --*)
                    fl="$section"
                    if [[ "$fix" == --test ]]; then title="To fix with $fl"; else title="Fixed by $fl"; fi
                    mapfile -t lines < <(printf '%s' "${changed[$fl]:-}")
                    ;;
            esac
            (( ${#lines[@]} )) || continue
            printf '\n## %s\n\n' "$title"
            # escape < so tags like <releasedate> are shown, not swallowed as HTML
            for m in "${lines[@]}"; do printf -- '- %s\n' "${m//</\\<}"; done
        done
    } > "$report"
}
trap write_report EXIT

[[ -f "$recipe" ]] || { err "no $recipe"; exit 1; }
# genres, port.json attrs, arches, banned/common/CFW-provided libraries
rules="$(dirname "${BASH_SOURCE[0]}")/port_rules.json"
jq -e . "$rules" > /dev/null 2>&1 || { err "$rules is missing or not valid JSON"; exit 1; }
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

# attr keys, genres and arches against port_rules.json
while IFS= read -r l; do err "$l"; done < <(jq -r --slurpfile r "$rules" '
    $r[0] as $r | .attr as $a
    | ($r.port_json_attrs[] | select(. as $k | $a | has($k) | not) | "port.json attr.\(.) is missing"),
      (if ($a.genres | type) != "array" or ($a.genres | length) == 0 then "port.json attr.genres must be a non-empty array"
       else $a.genres[] | select(. as $g | $r.genres | index($g) | not)
            | "port.json genre \"\(.)\" is not one of: \($r.genres | join(", "))" end),
      (if ($a.arch | type) != "array" or ($a.arch | length) == 0 then "port.json attr.arch must be a non-empty array"
       else $a.arch[] | select(. as $x | $r.arches | index($x) | not)
            | "port.json arch \"\(.)\" is not one of: \($r.arches | join(", "))" end)' "$pj")
while IFS= read -r k; do warn "port.json attr.$k is not a PortMaster field"; done \
    < <(jq -r --slurpfile r "$rules" '.attr | keys[] | select(. as $k | $r[0].port_json_attrs | index($k) | not)' "$pj")

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
    if [[ "$(git ls-files -s -- "$sh" | cut -c1-6)" == 100755 ]]; then
        if fixing; then
            fixit --fix "set $launcher to 644 (staged in git)" set_mode_644 "$sh"
        else
            warn "$launcher is committed as 755 (PortMaster uses 644; or run with --fix)"
        fi
    fi
    if command -v shellcheck > /dev/null; then
        # SC2188: "> file" with no command, e.g. the template's log truncation; ": > file" is the same
        if fixing; then
            while IFS=: read -r _ line _; do
                fixit --fix "$launcher:$line: change \"> file\" to \": > file\" (SC2188)" sed -i "${line}s/^\([[:space:]]*\)>/\1: >/" "$sh"
            done < <(shellcheck -s bash -i SC2188 -f gcc "$sh")
        fi
        while IFS= read -r l; do err "shellcheck $l"; done < <(shellcheck -s bash -S error -f gcc "$sh")
        # SC1090/1091 sourced files, SC2154 variables from control.txt, SC2164 cd without || exit;
        # SC2188 is listed above as a repair when fixing
        while IFS= read -r l; do warn "shellcheck $l"; done \
            < <(shellcheck -s bash -S warning -e "SC1090,SC1091,SC2154,SC2164,SC2034${fix:+,SC2188}" -f gcc "$sh")
    else
        note "shellcheck not installed, launcher lint skipped"
    fi
else
    err "launcher $launcher not found"
fi

# ---------------------------------------------------------------- binaries and bundled libraries
# Lists in port_rules.json: banned_libs (the CFW ships a device-specific build; error),
# gl_libs (only inside gl4es.<arch>/), common_libs (already on CFWs; warning) and
# cfw_provided (NEEDED entries that need not be bundled).
is_elf() { [[ -f "$1" && "$(head -c 4 "$1" | od -An -c | tr -d ' ')" == '177ELF' ]]; }
mapfile -t banned < <(jq -r '.banned_libs[] | "\(.pattern)\t\(.reason)"' "$rules")
gl_re=$(jq -r .gl_libs "$rules")
common_re=$(jq -r '.common_libs | join("|")' "$rules")
provided_re=$(jq -r '.cfw_provided | join("|")' "$rules")
arches_re="^($(jq -r '.arches | join("|")' "$rules"))$"
declare -A machine_arch=() bundled=() unresolved=()
libfiles=()   # bundled .so files kept (for the license check)
while IFS=$'\t' read -r m a; do machine_arch[$m]=$a; done \
    < <(jq -r '.elf_machine_arch | to_entries[] | "\(.key)\t\(.value)"' "$rules")
port_arches=" $(jq -r '.attr.arch // [] | join(" ")' "$pj") "
elfs=()
while IFS= read -r -d '' f; do
    rel="${f#"$dir"/}" name="${f##*/}" drop=""
    if [[ "$name" == *.so || "$name" == *.so.* ]]; then
        for b in "${banned[@]}"; do
            [[ "$name" =~ ${b%%$'\t'*} ]] && err "bundles $rel: ${b#*$'\t'}"
        done
        [[ "$name" =~ $gl_re && "$rel" != gl4es.*/* ]] && err "bundles $rel: libGL/libEGL only belong in a gl4es.<arch>/ folder"
        if [[ "$name" =~ $common_re ]]; then
            # --test lists the removal too (as SC2188 above), so it is not also a warning
            if [[ -n "$bloat" || "$fix" == --test ]]; then
                fixit --fix-bloat "remove $rel: CFWs already ship it (also drop it from FILES in build.sh)" rm -f "$f"
                [[ "$fix" == --test ]] || continue
                drop=1   # --test: still checked as a binary, but needs no license
            else
                warn "bundles $rel: CFWs already ship it, bundling only bloats the port (--fix-bloat removes it)"
            fi
        fi
        [[ -n "$drop" ]] || libfiles+=("$rel")
    fi
    is_elf "$f" || continue
    machine=$(readelf -h "$f" 2> /dev/null | sed -n 's/^ *Machine: *//p')
    arch="${machine_arch[$machine]:-}"
    [[ -n "$arch" ]] || { err "$rel is an ELF file for \"$machine\", which PortMaster does not run"; continue; }
    [[ "$port_arches" == *" $arch "* ]] || err "$rel is $arch but port.json attr.arch does not list $arch"
    # the arch named by a .<arch> suffix or a libs.<arch>/ or gl4es.<arch>/ folder must match the ELF
    tag=""
    if [[ "$rel" =~ (^|/)(libs|gl4es)\.([a-z0-9_]+)/ ]]; then
        dirtag="${BASH_REMATCH[3]}"
        [[ "$dirtag" =~ $arches_re ]] && tag="$dirtag"
    fi
    [[ -z "$tag" && "${name##*.}" =~ $arches_re ]] && tag="${name##*.}"
    [[ -n "$tag" && "$tag" != "$arch" ]] && err "$rel is named for $tag but is a $arch binary"
    bundled[$arch/$name]=1
    elfs+=("$arch"$'\t'"$f")
done < <(find "$dir" \( -type f -o -type l \) -print0)

# every bundled library needs licenses/LICENSE.<component>.txt. The component is the lib name
# without lib/.so/version (libSDL2_image-2.0.so.0 -> SDL2_image); it matches a license when
# one name starts with the other (libvorbisfile -> LICENSE.libvorbis.txt), case and
# punctuation ignored, or through license_aliases in port_rules.json (libcrypto -> openssl).
if [[ -d "$dir/licenses" ]]; then
    norm() { local s="${1,,}"; s="${s#lib}"; echo "${s//[^a-z0-9]/}"; }
    declare -A lic_alias=() lic_name=() lic_missing=()
    while IFS=$'\t' read -r k v; do lic_alias[$(norm "$k")]=$(norm "$v"); lic_name[$(norm "$k")]="$v"; done \
        < <(jq -r '.license_aliases // {} | to_entries[] | "\(.key)\t\(.value)"' "$rules")
    lics=()
    while IFS= read -r l; do lics+=("$(norm "$l")"); done \
        < <(find "$dir/licenses" -maxdepth 1 -type f -name 'LICENSE.*.txt' -printf '%f\n' | sed 's/^LICENSE\.\(.*\)\.txt$/\1/')
    for rel in "${libfiles[@]}"; do
        comp="${rel##*/}"; comp="${comp%%.so*}"; comp="${comp#lib}"
        [[ "$comp" =~ ^(.*[^0-9.])-[0-9][0-9.]*$ ]] && comp="${BASH_REMATCH[1]}"
        c=$(norm "$comp"); want_lic="${lic_name[$c]:-$comp}"; c="${lic_alias[$c]:-$c}"
        ok=""
        for l in "${lics[@]}"; do
            if [[ "$c" == "$l" ]] || { (( ${#l} >= 3 )) && [[ "$c" == "$l"* ]]; } || { (( ${#c} >= 3 )) && [[ "$l" == "$c"* ]]; }; then
                ok=1; break
            fi
        done
        [[ -n "$ok" ]] || lic_missing[$want_lic]+=" $rel"
    done
    # license_urls (name -> URL), matched the same way as license file names
    lic_match() { [[ "$1" == "$2" ]] || { (( ${#2} >= 3 )) && [[ "$1" == "$2"* ]]; } || { (( ${#1} >= 3 )) && [[ "$2" == "$1"* ]]; }; }
    mapfile -t lic_urls < <(jq -r '.license_urls // {} | to_entries[] | "\(.key)\t\(.value)"' "$rules")
    fetch_license() {   # url dest: refuse empty files and HTML error pages
        local tmp; tmp=$(mktemp)
        if curl -fsSL --max-time 30 "$1" -o "$tmp" && [[ -s "$tmp" ]] && ! head -c 512 "$tmp" | grep -qiE '<!doctype|<html'; then
            cp "$tmp" "$2" && chmod 644 "$2"; rm -f "$tmp"
        else
            rm -f "$tmp"; return 1
        fi
    }
    for k in "${!lic_missing[@]}"; do
        c=$(norm "$k") name="" url=""
        for e in "${lic_urls[@]}"; do
            lic_match "$c" "$(norm "${e%%$'\t'*}")" && { name="${e%%$'\t'*}"; url="${e#*$'\t'}"; break; }
        done
        if [[ -n "$url" && ( -n "$license" || "$fix" == --test ) ]]; then
            fixit --fix-license "fetch licenses/LICENSE.$name.txt for${lic_missing[$k]} from $url" \
                fetch_license "$url" "$dir/licenses/LICENSE.$name.txt" \
                || warn "could not download $url: add licenses/LICENSE.$name.txt for${lic_missing[$k]} by hand"
        else
            warn "no license for bundled${lic_missing[$k]}: add licenses/LICENSE.${name:-$k}.txt${license:+ (no URL in license_urls)}${url:+ (--fix-license fetches it)}"
        fi
    done
fi

# every NEEDED library should be bundled for that arch or be one the CFWs provide
for e in "${elfs[@]}"; do
    arch="${e%%$'\t'*}" f="${e#*$'\t'}"
    while IFS= read -r n; do
        [[ "$n" =~ $provided_re || -n "${bundled[$arch/$n]:-}" ]] || unresolved[$arch/$n]+=" ${f#"$dir"/}"
    done < <(readelf -d "$f" 2> /dev/null | sed -n 's/.*(NEEDED).*\[\(.*\)\]/\1/p')
done
for k in "${!unresolved[@]}"; do
    # a warning: many CFWs have more than cfw_provided lists; test_port.sh's ldd on devices is the proof
    warn "${k#*/} (${k%%/*}) is needed by${unresolved[$k]} but not bundled or in cfw_provided: check ldd on devices"
done

for f in "$dir"/*.aarch64 "$dir"/*.x86_64; do
    is_elf "$f" || continue
    readelf -W --dyn-syms "$f" 2> /dev/null | awk '$7 != "UND" && $8 == "SDL_Init"' | grep -q . \
        && err "${f#"$dir"/} has SDL2 linked in statically: link against the system libSDL2"
done

# ---------------------------------------------------------------- glibc
# min_glibc must be the highest GLIBC_x.y any aarch64 binary or bundled lib asks for (blank
# when that is 2.17, the first aarch64 glibc): PortMaster hides the port on older devices.
vle() { [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n 1)" == "$1" ]]; }
need="" need_file=""
while IFS= read -r f; do
    is_elf "$f" || continue
    v=$(readelf -W --version-info "$f" 2> /dev/null | grep -o 'GLIBC_[0-9][0-9.]*' | sed 's/GLIBC_//' | sort -uV | tail -n 1)
    [[ -n "$v" ]] && { [[ -z "$need" ]] || ! vle "$v" "$need"; } && { need="$v"; need_file="${f#"$dir"/}"; }
done < <(printf '%s\n' "$dir"/*.aarch64; find "$dir/libs.aarch64" -type f 2> /dev/null)
min=$(jq -r '.attr.min_glibc // ""' "$pj")
if [[ -z "$need" ]]; then
    jq -e '.attr.arch | index("aarch64")' "$pj" > /dev/null && warn "no aarch64 binaries in $dir (not built yet?)"
else
    want="$need"; vle "$need" 2.17 && want=""
    if [[ "$min" != "$want" ]] && fixing; then
        fixit --fix "set attr.min_glibc from \"$min\" to \"$want\" ($need_file needs GLIBC_$need)" set_min_glibc "$want"
    elif [[ "$min" != "$want" ]]; then
        if [[ -n "$min" ]] && vle "$need" "$min"; then
            warn "attr.min_glibc is $min but the aarch64 binaries only need ${want:-nothing above 2.17} (devices in between are excluded)"
        else
            err "attr.min_glibc is ${min:-empty} but $need_file needs GLIBC_$need: set it to \"$want\" (or run with --fix)"
        fi
    fi
fi

# ---------------------------------------------------------------- runtimes
# love_11.5 ships inside PortMaster itself (PortMaster/runtimes/), so it is not a runtimes.json key
builtin_runtimes='love_11.5'
mapfile -t runtimes < <(jq -r '.attr.runtime // [] | if type == "array" then .[] else "NOT_AN_ARRAY" end' "$pj")
if (( ${#runtimes[@]} )); then
    rj=$(curl -sfL --max-time 20 https://raw.githubusercontent.com/PortsMaster/PortMaster-New/main/runtimes/runtimes.json)
    for r in "${runtimes[@]}"; do
        if [[ "$r" == NOT_AN_ARRAY ]]; then err "attr.runtime must be a JSON array"
        elif [[ " $builtin_runtimes " == *" $r "* ]] && fixing; then
            fixit --fix "remove $r from attr.runtime (it ships with PortMaster)" drop_runtime "$r"
        elif [[ " $builtin_runtimes " == *" $r "* ]]; then warn "attr.runtime lists $r, which ships with PortMaster and is not in runtimes.json: remove it (or run with --fix)"
        elif [[ -z "$rj" ]]; then warn "could not download runtimes.json, attr.runtime $r not checked"
        elif ! jq -e --arg r "$r" 'has($r)' <<< "$rj" > /dev/null; then err "attr.runtime $r is not a key in PortMaster's runtimes.json"
        fi
    done
fi

# ---------------------------------------------------------------- gameinfo.xml
gi="$dir/gameinfo.xml"
if [[ -f "$gi" ]]; then
    while IFS= read -r l; do err "$l"; done < <(python3 -I - "$gi" "$root" "./$launcher" "$(jq -r '.gameinfo_tags | join(" ")' "$rules")" <<'EOF'
import os, re, sys, xml.etree.ElementTree as ET
gi, root, launcher, tags = sys.argv[1:]
try:
    top = ET.parse(gi).getroot()
except ET.ParseError as e:
    sys.exit(print("gameinfo.xml is not valid XML: %s" % e))
if top.tag != "gameList":
    print("gameinfo.xml root element is <%s>, it must be <gameList>" % top.tag)
games = top.findall("game")
for g in games:
    if (g.findtext("path") or "") != launcher:
        print("gameinfo.xml <path> is %r, the launcher is %r" % (g.findtext("path"), launcher))
    missing = [t for t in tags.split() if not (g.findtext(t) or "").strip()]
    if missing:
        print("gameinfo.xml is missing <%s>" % ">, <".join(missing))
    date = (g.findtext("releasedate") or "").strip()
    if date and not re.fullmatch(r"\d{8}T\d{6}", date):
        print("gameinfo.xml <releasedate> is %r, use YYYYMMDDT000000" % date)
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
    # --fix renames LICENSE_x(.txt|.md) / LICENSE-x / COPYING_x to LICENSE.x.txt; other names need a component
    odd=""
    while IFS= read -r l; do
        new=""
        [[ "$l" =~ ^(LICEN[SC]E|COPYING)[_-](.+)$ ]] && new="LICENSE.${BASH_REMATCH[2]%.*}.txt"
        [[ -n "$new" && "${BASH_REMATCH[2]}" != *.* ]] && new="LICENSE.${BASH_REMATCH[2]}.txt"
        if fixing && [[ -n "$new" && ! -e "$dir/licenses/$new" ]]; then
            fixit --fix "rename licenses/$l to $new" mv "$dir/licenses/$l" "$dir/licenses/$new"
        else
            odd+="$l "
        fi
    done < <(find "$dir/licenses" -maxdepth 1 -type f ! -name 'LICENSE.*.txt' ! -name .gitkeep -printf '%f\n')
    [[ -n "$odd" ]] && warn "licenses/ names should be LICENSE.<component>.txt: $odd"
    find "$dir/licenses" -maxdepth 1 -type f ! -name .gitkeep | grep -q . || err "licenses/ is empty"
else
    err "no licenses/ folder"
fi

# ---------------------------------------------------------------- screenshot and cover
# PortMaster wants both under 100 kB and the screenshot 640x480. --fix letterboxes the
# screenshot to 640x480 and re-encodes oversized images with fewer palette colours
# (pngquant if installed, else ImageMagick) until they fit.
max_bytes=100000
png_size() { od -An -tu1 -j16 -N8 "$1" | awk '{printf "%dx%d", $1*16777216+$2*65536+$3*256+$4, $5*16777216+$6*65536+$7*256+$8}'; }
im=$(command -v magick || command -v convert || true)
# the trial encodes go to a temp file outside the port; --test only reports the size reached
shrink_png() {
    local f="$1" tmp c
    tmp=$(mktemp --suffix=.png)
    for c in 256 128 64 32 16; do
        if command -v pngquant > /dev/null; then
            pngquant --force --strip --speed 1 --output "$tmp" "$c" -- "$f" || continue
        elif [[ -n "$im" ]]; then
            "$im" "$f" -strip +dither -colors "$c" -define png:compression-level=9 "PNG8:$tmp" || continue
        else
            rm -f "$tmp"; note "pngquant or ImageMagick is needed to shrink ${f#"$dir"/}"; return 1
        fi
        if (( $(stat -c%s "$tmp") < max_bytes )); then
            # cp keeps the PNG's own mode (mktemp files are 600)
            fixit --fix "shrink ${f#"$dir"/} from $(( $(stat -c%s "$f") / 1024 )) to $(( $(stat -c%s "$tmp") / 1024 )) kB ($c colours)" cp "$tmp" "$f"
            rm -f "$tmp"; return 0
        fi
    done
    rm -f "$tmp"
    return 1
}
for img in screenshot.png cover.png; do
    f="$dir/$img"
    if [[ ! -f "$f" ]]; then warn "no $img (make_images.sh creates a placeholder)"; continue; fi
    if [[ "$img" == screenshot.png && "$(png_size "$f")" != 640x480 ]]; then
        if fixing && [[ -n "$im" ]]; then
            fixit --fix "letterbox $img from $(png_size "$f") to 640x480" \
                "$im" "$f" -resize 640x480 -background black -gravity center -extent 640x480 "$f"
        else
            err "$img is $(png_size "$f"), it must be 640x480 (or run with --fix)"
        fi
    fi
    if (( $(stat -c%s "$f") >= max_bytes )); then
        if fixing; then
            shrink_png "$f" || err "$img is still $(( $(stat -c%s "$f") / 1024 )) kB after compression, it must be under 100 kB"
        else
            err "$img is $(( $(stat -c%s "$f") / 1024 )) kB, it must be under 100 kB (or run with --fix)"
        fi
    fi
    if grep -qa 'PortMaster placeholder' "$f"; then
        if [[ "$img" == screenshot.png ]] && jq -e .screenshot "$recipe" > /dev/null; then
            warn "$img is still the placeholder although recipe.json has a screenshot config: the Xvfb capture failed (see Capture screenshot)"
        else
            warn "$img is a generated placeholder: replace it before submitting (test_port.sh --screenshot DEV)"
        fi
    fi
done

if [[ "$fix" == --test ]]; then
    for fl in "${fix_flags[@]}"; do
        [[ -n "${changed[$fl]:-}" ]] || continue
        echo "check_port: $port: to fix with $fl:"
        printf '%s' "${changed[$fl]}" | sed 's/^/    /'
    done
    echo "check_port: $port: --test: changes: $(change_counts); $errors error(s) and ${#warns[@]} warning(s) need a manual fix"
    (( errors )) && exit 1
    exit 0
fi
if (( errors )); then
    echo "check_port: $port: $errors error(s)"
    exit 1
fi
echo "check_port: $port: ok"
