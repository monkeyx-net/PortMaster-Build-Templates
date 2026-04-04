# contrsol4PM

A Love2D tool for PortMaster with two modes:

1. **Viewer** — displays a control map image fullscreen with an auto-exit countdown
2. **Creator** — generates a control map SVG/PNG from a game's README and `.gptk` config file

## Requirements

- [Love2D](https://love2d.org/) 11.4 or later
- [Inkscape](https://inkscape.org/) (required by the creator mode to export PNG)

---

## Viewer Mode

Displays an image fullscreen. Exits after a timeout or on any key/button press.

```
love . [image_path] [duration_seconds] [width] [height]
```

### Arguments

| Argument | Default | Description |
|---|---|---|
| `image_path` | `controls.png` | Path to the image to display |
| `duration_seconds` | `5` | Seconds before auto-exit |
| `width` | image width | Scale image to this pixel width |
| `height` | image height | Scale image to this pixel height |

All arguments are optional and positional. To pass `width`/`height` you must also pass `duration_seconds`.

### Examples

```bash
# Display controls.png for 5 seconds (defaults)
love .

# Custom image
love . edgar.png

# Custom image, 10 second timeout
love . edgar.png 10

# Custom image, 5 second timeout, scaled to 640x480
love . edgar.png 5 640 480

# Scale width only (height stays natural)
love . edgar.png 5 640

# Show help
love . /?
```

### Exit Conditions

- Duration elapses
- Any keyboard key is pressed
- Any gamepad button is pressed
- Window is closed

---

## Creator Mode

Generates a control map image from a game's README and `.gptk` keyboard mapping file.

```
love . create -i INPUT_SVG -r README_FILE [-c CONFIG_FILE]
```

### How It Works

1. Reads button action descriptions from the game's `README.md` (markdown table format)
2. Reads keyboard mappings from the `.gptk` config file (optional)
3. Substitutes placeholders in the SVG template with the mapped values
4. README values take priority; `.gptk` values are used as fallback
5. Long text (>15 chars) is automatically line-wrapped in the SVG
6. Exports the result as both `.svg` and `.png` via Inkscape

### Options

| Option | Description |
|---|---|
| `-i`, `--input FILE` | SVG template file containing placeholders |
| `-r`, `--readme FILE` | Game README.md with a controls table |
| `-c`, `--config FILE` | `.gptk` keyboard mapping file (optional) |
| `-p`, `--port FILE` | `port.json` file to read game title from (optional, overrides `-c` title) |
| `-h`, `--help` | Show help |

### SVG Placeholders

The SVG template uses Z-prefixed placeholders that get replaced:

| Placeholder | Replaced with |
|---|---|
| `ZAD1` | A button action (from README) |
| `ZBD1` | B button action |
| `ZXD1` | X button action |
| `ZYD1` | Y button action |
| `ZL1D1` | L1 button action |
| `ZL2D1` | L2 button action |
| `ZR1D1` | R1 button action |
| `ZR2D1` | R2 button action |
| `ZA` `ZB` `ZX` `ZY` | Keyboard key for each button (from `.gptk`) |
| `ZL1` `ZL2` `ZR1` `ZR2` | Keyboard key for shoulder buttons |
| `ZUP` `ZDOWN` `ZLEFT` `ZRIGHT` | D-pad keyboard keys (defaults: Up/Down/Left/Right) |
| `ZSELECT` | Select/Back key |
| `ZSTART` | Start key |
| `ZTITLE` | Game title (from `-p` port.json, or derived from `-c` config filename) |

### README Format

The game README must contain a markdown controls table:

```markdown
| Button | Action |
|--|--|
| A | Jump |
| B | Inventory |
| L1 | Attack |
```

Button names are case-insensitive (`A`, `a`, `Select`, `SELECT` all work).

### Examples

```bash
# Generate from README + gptk config + port.json title
love . create -p port.json -c edgar.gptk -i templates/PM_MAPPINGS_TEMPLATE.svg -r /path/to/README.md

# Generate using README only (no gptk config)
love . create -i templates/PM_MAPPINGS_TEMPLATE.svg -r /path/to/README.md

# Show creator help
love . create -h
```

Output files are always written to the current directory as:
- `controls.svg`
- `controls.png`
