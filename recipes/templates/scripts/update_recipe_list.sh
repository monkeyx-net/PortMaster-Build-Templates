#!/bin/bash
# update_recipe_list.sh
#
# Regenerates the recipe_file dropdown in .github/workflows/build_recipe.yml
# with the port folder names from recipes/ports/*/recipe.json. GitHub reads workflow_dispatch choice options
# statically from the YAML, so the list has to be written into the file.
#
# Replaces the lines between the recipe-list-start and recipe-list-end markers.
# Exits non-zero without touching the file if a marker is missing or out of
# order, or if no recipes are found.
#
# Usage (from the repo root):
#   recipes/templates/scripts/update_recipe_list.sh

set -euo pipefail
shopt -s nullglob
export LC_ALL=C

WORKFLOW=".github/workflows/build_recipe.yml"
START_RE='^ *# recipe-list-start'
END_RE='^ *# recipe-list-end'

[[ -f "$WORKFLOW" ]] || { echo "File not found: $WORKFLOW"; exit 1; }

recipes=(recipes/ports/*/recipe.json)
[[ ${#recipes[@]} -gt 0 ]] || { echo "No recipes found in recipes/ports"; exit 1; }

awk -v start="$START_RE" -v end="$END_RE" '
    $0 ~ start { starts++; if (!start_line) start_line=NR }
    $0 ~ end   { ends++;   if (!end_line)   end_line=NR }
    END { exit !(starts == 1 && ends == 1 && start_line < end_line) }
' "$WORKFLOW" || {
    echo "Expected one recipe-list-start marker followed by one recipe-list-end marker in $WORKFLOW"
    exit 1
}

tmp_list=$(mktemp)
trap 'rm -f "$tmp_list" "$WORKFLOW.tmp"' EXIT

for r in "${recipes[@]}"; do d=${r%/recipe.json}; printf "          - '%s'\n" "${d##*/}"; done | sort > "$tmp_list"

awk -v start="$START_RE" -v end="$END_RE" -v listfile="$tmp_list" '
    $0 ~ start {
        print
        while ((getline line < listfile) > 0) print line
        skip=1; next
    }
    $0 ~ end { skip=0 }
    !skip { print }
' "$WORKFLOW" > "$WORKFLOW.tmp" && mv "$WORKFLOW.tmp" "$WORKFLOW"

echo "$WORKFLOW updated with ${#recipes[@]} recipes"
