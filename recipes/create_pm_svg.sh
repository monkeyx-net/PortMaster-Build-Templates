#!/bin/bash

extract_buttons_from_file() {
    local readme=$1
    unset button_mapping
    declare -gA button_mapping

    buttons=(A B X Y L1 L2 R1 R2 SELECT START)
    for button in "${buttons[@]}"; do
        button_mapping["$button"]="n/a"
    done

    while IFS= read -r line; do
        if [[ "$line" =~ ^\|\ +([A-Z0-9]+)\ +\|\ +([^|]+)\ +\| ]]; then
            button="${BASH_REMATCH[1]}"
            action="${BASH_REMATCH[2]}"
            if [[ "$button" =~ ^(A|B|X|Y|L1|L2|R1|R2|SELECT|START)$ ]]; then
                action=$(echo "$action" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
                button_mapping["$button"]="$action"
                #echo "Button $button maps to: $action"
            fi
        fi
    done < "${readme}"
    echo "=== Button Mappings ==="
    for button in "${buttons[@]}"; do
        printf "%-3s : %s\n" "$button" "${button_mapping[$button]}"
    done
    export buttons
    export button_mapping
    return 0
}

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo "Replace Z-pattern placeholders in an SVG file with values from a config file."
    echo ""
    echo "Options:"
    echo "  -c, --config FILE    Config file to read key-value pairs gptk file"
    echo "  -i, --input FILE     Input SVG file to modify"
    echo "  -h, --help           Show this help message"
    echo "  -r, --readme         The README.md file for a port"
    echo ""
    echo "Examples:"
    echo "  $0 -c mygame.gptk -i template.svg -o result.svg -r /home/user/port/README.md"
    echo "  $0 --config controls.cfg --input input.svg --reaadme /home/user/port/README.md"
    echo "  $0 -h"
    exit 0
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -i|--input)
            SVG_FILE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            ;;
        -r|--readme)
            README_FILE="$2"
            shift 2
            ;;
        *)
            echo "Error: Unknown option $1"
            echo "Use $0 -h for help"
            exit 1
            ;;
    esac
done

# Check if files exist
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config file $CONFIG_FILE not found!"
    exit 1
fi
if [ ! -f "$SVG_FILE" ]; then
    echo "Error: SVG file $SVG_FILE not found!"
    exit 1
fi
if [ ! -f "$README_FILE" ]; then
    echo "Error: SVG file $SVG_FILE not found!"
    exit 1
fi

CONFIG_BASENAME=$(basename "$CONFIG_FILE" .gptk)
# First letter uppercase
TITLE="${CONFIG_BASENAME^}"
content=$(cat "$SVG_FILE")
OUTPUT_FILE=$CONFIG_BASENAME.svg

extract_buttons_from_file $README_FILE
for button in "${buttons[@]}"; do
    case $button in
         L1|L2|R1|R2|A|B|X|Y|SELECT|START)
            if [[ ${#button} -eq 1 ]]; then
                # Single-character buttons: A, B, X, Y
                placeholder="Z${button}D1"
            elif [[ ${#button} -eq 2 ]]; then
                # Two-character buttons: L1, L2, R1, R2
                placeholder="Z${button:0:1}D${button:1:1}"
            else
                # Longer buttons: SELECT, START
                placeholder="Z${button}"
            fi
            replacement="${button_mapping[$button]//\//\\/}"
            content=$(echo "$content" | sed "s/$placeholder/$replacement/g")
            ;;
    esac
done

while IFS='=' read -r key value; do
    key=$(echo "$key" | xargs)
    value=$(echo "$value" | xargs)
    if [[ -z "$key" || -z "$value" ]]; then
        continue
    fi
    zkey=$(echo "$key" | tr '[:lower:]' '[:upper:]')
    case $key in
        "select")
            content=$(echo "$content" | sed "s/ZSELECT/${zkey}/g")
            content=$(echo "$content" | sed "s/ZTITLE/$TITLE/g")
            ;;
        "start")
            content=$(echo "$content" | sed "s/ZSTART/${zkey}/g")
            ;;
        "a"|"b"|"x"|"y"|"up"|"down"|"left"|"right"|"l1"|"l2"|"r1"|"r2")
            content=$(echo "$content" | sed "s/Z${zkey}/${zkey}/g")
           #content=$(echo "$content" | sed "s/${zkey} - Z${zkey} -/${zkey} - ${value} -/g")
            ;;
    esac
done < "$CONFIG_FILE"

echo "$content" > "$OUTPUT_FILE"

inkscape $OUTPUT_FILE -o $CONFIG_BASENAME.png

echo "Replacements complete!"
echo "Modified file saved as: $OUTPUT_FILE"
echo "PNG file created as: $CONFIG_BASENAME.png"
