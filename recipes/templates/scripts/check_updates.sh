#!/bin/bash
# check_updates.sh
#
# Based on templates/scripts/check.sh — applies the same upstream fetch logic
# to every port in recipes/ports/*/recipe.json and reports differences between
# the stored port_version / port_checksum and what's currently upstream.
#
# port_url formats supported (mirrors check.sh):
#   owner/repo                               GitHub repo; checks releases/latest,
#                                            falls back to default-branch archive
#   https://github.com/…/refs/tags/V.ext    Full GitHub tag URL
#   https://github.com/…/refs/heads/B.ext   Full GitHub branch URL
#   https://codeberg.org/owner/repo/…       Codeberg; checks branch API
#   https://…                               Any other direct URL
#
# Status codes:
#   [OK]     stored version & checksum match upstream
#   [INFO]   upstream resolved but recipe fields not yet populated
#   [UPDATE] stored value differs from upstream — recipe needs updating
#   [ERROR]  download or API call failed
#
# Usage:
#   ./check_updates.sh              # check all ports
#   ./check_updates.sh soh shamogu  # check specific ports

set -uo pipefail
CURRENT_DIR="$(pwd)"
RECIPES_DIR="$CURRENT_DIR/recipes/ports"
WORK_DIR="$(mktemp -d)"
DOWNLOADS_DIR="$CURRENT_DIR/new_ports"
PM_PORTS_JSON="$WORK_DIR/portmaster_ports.json"
PM_PORTS_URL="https://raw.githubusercontent.com/PortsMaster/PortMaster-Info/main/ports.json"

TOTAL=0; OK=0; UPDATES=0; ERRORS=0; SKIPPED=0

cleanup() { rm -rf "$WORK_DIR"; }
trap cleanup EXIT

for cmd in jq curl wget sha256sum; do
    command -v "$cmd" &>/dev/null || { echo "Missing required tool: $cmd"; exit 1; }
done

# ─── API helpers ─────────────────────────────────────────────────────────────

# Latest release tag for a GitHub repo, empty string if none.
gh_latest_release() {
    curl -sf "https://api.github.com/repos/$1/releases/latest" \
        | jq -r '.tag_name // empty' 2>/dev/null || true
}

# Default branch name for a GitHub repo.
gh_default_branch() {
    curl -sf "https://api.github.com/repos/$1" \
        | jq -r '.default_branch // empty' 2>/dev/null || true
}

# Short SHA of the HEAD commit on a GitHub branch.
gh_branch_sha() {
    curl -sf "https://api.github.com/repos/$1/commits/$2" \
        | jq -r '.sha // empty' 2>/dev/null | cut -c1-7 || true
}

# Short SHA of the HEAD commit on a Codeberg branch (Gitea API).
cb_branch_sha() {                            # args: owner repo branch
    curl -sf "https://codeberg.org/api/v1/repos/$1/$2/branches/$3" \
        | jq -r '.commit.id // empty' 2>/dev/null | cut -c1-7 || true
}

# Latest release tag for a Codeberg repo, empty string if none.
cb_latest_release() {                        # args: owner repo
    local tag
    tag=$(curl -sf "https://codeberg.org/api/v1/repos/$1/$2/releases/latest" \
        | jq -r '.tag_name // empty' 2>/dev/null) || tag=''
    if [[ -z "$tag" || "$tag" == 'null' ]]; then
        tag=$(curl -sf "https://codeberg.org/api/v1/repos/$1/$2/tags?limit=1" \
            | jq -r '.[0].name // empty' 2>/dev/null) || tag=''
    fi
    echo "$tag"
}

# Download URL to file; return 1 on failure.
fetch() { wget -q -O "$2" "$1" 2>/dev/null; }

# Look up a port zip name in the cached PortMaster ports.json, download it
# to dest_dir using the source.url field, and print the result.
pm_zip_download() {
    local zip_name="$1" dest_dir="$2"
    local pm_url
    pm_url=$(jq -r --arg n "$zip_name" '.ports[$n].source.url // empty' "$PM_PORTS_JSON" 2>/dev/null)
    if [[ -z "$pm_url" ]]; then
        echo "         portmaster: $zip_name not found in ports.json"
        return
    fi
    if fetch "$pm_url" "$dest_dir/$zip_name"; then
        echo "         portmaster: saved -> $dest_dir/$zip_name"
        mkdir -p "$dest_dir/$port_name"
        bsdtar -xf "$dest_dir/$zip_name" -C "$dest_dir/$port_name/"
    else
        echo "         portmaster: download failed -- $pm_url"
    fi
}

# ─── check one port ──────────────────────────────────────────────────────────
check_port() {
    local port_name="$1"
    local recipe="$RECIPES_DIR/$port_name/recipe.json"

    [[ -f "$recipe" ]] || {
        echo "[ERROR]  $port_name -- recipe.json not found"
        ERRORS=$((ERRORS+1)); return
    }
    TOTAL=$((TOTAL+1))

    local port_zip port_url stored_version stored_checksum
    port_zip=$(jq -r '.name // empty' "$recipe")
    port_url=$(jq -r '.source.port_url // empty' "$recipe")
    stored_version=$(jq -r '.source.port_version // empty' "$recipe")
    stored_checksum=$(jq -r '.source.port_checksum // empty' "$recipe")

    if [[ -z "$port_url" ]]; then
        echo "[SKIP]   $port_name -- port_url is empty"
        SKIPPED=$((SKIPPED+1)); return
    fi

    local tmpdir="$WORK_DIR/$port_name"
    mkdir -p "$tmpdir"

    local upstream_version="" upstream_checksum="" fetch_error=false

    # ── (1) GitHub repo path — owner/repo, no protocol ───────────────────────
    if [[ ! "$port_url" =~ ^https?:// ]]; then
        local repo="$port_url"
        local tag
        tag=$(gh_latest_release "$repo")

        if [[ -n "$tag" && "$tag" != "null" ]]; then
            upstream_version="$tag"
            fetch "https://github.com/${repo}/archive/refs/tags/${tag}.tar.gz" \
                  "$tmpdir/source.tar.gz" \
                || { fetch_error=true; }
            [[ "$fetch_error" == false ]] \
                && upstream_checksum=$(sha256sum "$tmpdir/source.tar.gz" | cut -d' ' -f1)
        else
            local branch
            branch=$(gh_default_branch "$repo")
            if [[ -z "$branch" ]]; then
                fetch_error=true
            else
                local sha
                sha=$(gh_branch_sha "$repo" "$branch")
                [[ -n "$sha" ]] && upstream_version="commit:$sha"
                fetch "https://github.com/${repo}/archive/refs/heads/${branch}.zip" \
                      "$tmpdir/source.zip" \
                    || { fetch_error=true; }
                [[ "$fetch_error" == false ]] \
                    && upstream_checksum=$(sha256sum "$tmpdir/source.zip" | cut -d' ' -f1)
            fi
        fi

    # ── (2) Full GitHub tag URL ───────────────────────────────────────────────
    # e.g. https://github.com/Owner/Repo/archive/refs/tags/v1.2.3.zip
    elif [[ "$port_url" =~ ^https://github\.com/([^/]+/[^/]+)/archive/refs/tags/([^/]+)\.(zip|tar\.gz)$ ]]; then
        local repo="${BASH_REMATCH[1]}"
        local url_tag="${BASH_REMATCH[2]}"
        local ext="${BASH_REMATCH[3]}"

        # Use the tag embedded in the URL as the stored baseline when field is empty
        [[ -z "$stored_version" ]] && stored_version="$url_tag"

        # Check if a newer release exists
        local latest_tag
        latest_tag=$(gh_latest_release "$repo")
        [[ -z "$latest_tag" || "$latest_tag" == "null" ]] \
            && latest_tag=$(curl -sf "https://api.github.com/repos/${repo}/tags" \
                | jq -r '.[0].name // empty' 2>/dev/null || true)
        [[ -n "$latest_tag" ]] && upstream_version="$latest_tag"

        # Verify checksum against the URL that's actually stored (not the latest)
        fetch "$port_url" "$tmpdir/source.$ext" || { fetch_error=true; }
        [[ "$fetch_error" == false ]] \
            && upstream_checksum=$(sha256sum "$tmpdir/source.$ext" | cut -d' ' -f1)

    # ── (3) Full GitHub branch URL ────────────────────────────────────────────
    # e.g. https://github.com/Owner/Repo/archive/refs/heads/main.zip
    elif [[ "$port_url" =~ ^https://github\.com/([^/]+/[^/]+)/archive/refs/heads/([^/]+)\.(zip|tar\.gz)$ ]]; then
        local repo="${BASH_REMATCH[1]}"
        local branch="${BASH_REMATCH[2]}"
        local ext="${BASH_REMATCH[3]}"

        local sha
        sha=$(gh_branch_sha "$repo" "$branch")
        [[ -n "$sha" ]] && upstream_version="commit:$sha"

        fetch "$port_url" "$tmpdir/source.$ext" || { fetch_error=true; }
        [[ "$fetch_error" == false ]] \
            && upstream_checksum=$(sha256sum "$tmpdir/source.$ext" | cut -d' ' -f1)

    # ── (4) Codeberg URL ──────────────────────────────────────────────────────
    # e.g. https://codeberg.org/owner/repo/archive/master.zip
    elif [[ "$port_url" =~ ^https://codeberg\.org/([^/]+)/([^/]+)/archive/([^.]+)\.(zip|tar\.gz)$ ]]; then
        local cb_owner="${BASH_REMATCH[1]}"
        local cb_repo="${BASH_REMATCH[2]}"
        local cb_ref="${BASH_REMATCH[3]}"
        local ext="${BASH_REMATCH[4]}"

        local release
        release=$(cb_latest_release "$cb_owner" "$cb_repo")
        if [[ -n "$release" && "$release" != 'null' ]]; then
            upstream_version="$release"
        else
            local sha
            sha=$(cb_branch_sha "$cb_owner" "$cb_repo" "$cb_ref")
            [[ -n "$sha" ]] && upstream_version="commit:$sha"
        fi

        fetch "$port_url" "$tmpdir/source.$ext" || { fetch_error=true; }
        [[ "$fetch_error" == false ]] \
            && upstream_checksum=$(sha256sum "$tmpdir/source.$ext" | cut -d' ' -f1)

    # ── (5) Any other direct URL ──────────────────────────────────────────────
    else
        local clean_url="${port_url%%\?*}"
        local ext="${clean_url##*.}"
        fetch "$port_url" "$tmpdir/source.$ext" || { fetch_error=true; }
        [[ "$fetch_error" == false ]] \
            && upstream_checksum=$(sha256sum "$tmpdir/source.$ext" | cut -d' ' -f1)
    fi

    # ── Compare stored vs upstream ────────────────────────────────────────────
    local version_diff=false  checksum_diff=false
    local version_unset=false checksum_unset=false

    if [[ -n "$upstream_version" ]]; then
        if [[ -z "$stored_version" ]]; then
            version_unset=true
        elif [[ "$upstream_version" != "$stored_version" ]]; then
            version_diff=true
        fi
    fi

    if [[ -n "$upstream_checksum" ]]; then
        if [[ -z "$stored_checksum" ]]; then
            checksum_unset=true
        elif [[ "$upstream_checksum" != "$stored_checksum" ]]; then
            checksum_diff=true
        fi
    fi

    # ── Report ────────────────────────────────────────────────────────────────
    if [[ "$fetch_error" == true ]]; then
        echo "[ERROR]  $port_name"
        echo "         $port_url"
        echo "         download or API call failed"
        ERRORS=$((ERRORS+1))

    elif [[ "$version_diff" == true || "$checksum_diff" == true || "$version_unset" == true || "$checksum_unset" == true ]]; then
        echo "[UPDATE] $port_name"
        UPDATES=$((UPDATES+1))
        echo "         $port_url"
        if [[ "$version_diff" == true ]]; then
            printf   "         version:  %-40s  (stored)\n"   "$stored_version"
            printf   "                   %-40s  (upstream)\n" "$upstream_version"
        elif [[ "$version_unset" == true ]]; then
            printf   "         version:  %-40s  (unset -- populate recipe)\n" "$upstream_version"
        fi
        if [[ "$checksum_diff" == true ]]; then
            printf   "         checksum: %s  (stored)\n"   "$stored_checksum"
            printf   "                   %s  (upstream)\n" "$upstream_checksum"
        elif [[ "$checksum_unset" == true ]]; then
            printf   "         checksum: %s  (unset -- populate recipe)\n" "$upstream_checksum"
        fi
        mkdir -p "$DOWNLOADS_DIR/$port_name/source"
        mv "$tmpdir"/source.* "$DOWNLOADS_DIR/$port_name/" 2>/dev/null || true
        echo "         saved -> $DOWNLOADS_DIR/$port_name/"
        _archives=("$DOWNLOADS_DIR/$port_name"/source.*)
        [[ -f "${_archives[0]}" ]] && bsdtar -xf "${_archives[0]}" -C "$DOWNLOADS_DIR/$port_name/source/" --strip-components=1
        [[ -n "$port_zip" ]] && pm_zip_download "$port_zip" "$DOWNLOADS_DIR/$port_name"
        local today
        today=$(date '+%Y-%m-%d')
        jq --arg v "$upstream_version" \
           --arg c "$upstream_checksum" \
           --arg d "$today" \
           '.source.port_version  = (if $v != "" then $v else .source.port_version  end) |
            .source.port_checksum = (if $c != "" then $c else .source.port_checksum end) |
            .source.date_updated  = $d' \
           "$recipe" > "$recipe.tmp" && mv "$recipe.tmp" "$recipe"
        echo "         recipe.json updated ($today)"
    else
        local detail=""
        [[ -n "$upstream_version" ]] && detail="  $upstream_version"
        echo "[OK]     $port_name$detail"
        OK=$((OK+1))
    fi
}

# ─── main ─────────────────────────────────────────────────────────────────────
echo "PortMaster Recipe Update Checker"
echo "=================================================="

[[ -d "$RECIPES_DIR" ]] || { echo "Directory not found: $RECIPES_DIR"; exit 1; }

echo -n "Fetching PortMaster ports.json ... "
if fetch "$PM_PORTS_URL" "$PM_PORTS_JSON"; then
    echo "done"
else
    echo "failed -- PortMaster zip lookup will be skipped"
    echo '{"ports":{}}' > "$PM_PORTS_JSON"
fi

# Default to all ports when no arguments given
if [[ $# -eq 0 ]]; then
    set -- "$RECIPES_DIR"/*/
    set -- "${@%/}"   # strip trailing slashes
    set -- "${@##*/}" # keep only basenames
fi

for port_name in "$@"; do
    check_port "$port_name"
done

echo "=================================================="
echo "Checked: $TOTAL  OK: $OK  Updates: $UPDATES  Errors: $ERRORS  Skipped: $SKIPPED"
