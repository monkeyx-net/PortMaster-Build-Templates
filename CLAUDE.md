# CLAUDE.md

This repo builds PortMaster ports with GitHub Actions (aarch64 + x86_64) and holds the metadata/assets for each port. Finished ports are submitted to [PortMaster-New](https://github.com/PortsMaster/PortMaster-New), whose reviewers reject verbose, badly built AI-generated ports without discussion. Read their [AGENTS.md](https://github.com/PortsMaster/PortMaster-New/blob/main/AGENTS.md) and follow it; the key rules are repeated below.

## Repo layout

```
recipes/
  new_recipe.sh                 # interactive: scaffolds a recipe + new_ports/<port>/ from templates
  ports/<port>/recipe.json      # source repo, build command, exe name, upstream version/checksum
  ports/<port>/build.sh         # runs inside the builder container (copied from templates/scripts/build.sh)
  templates/port_template/      # zz_port.sh launcher + zz_folder skeleton (port.json, gameinfo.xml, gptk, licenses)
  templates/includes/           # shared launcher includes (e.g. swap.txt), copied into ports listed in recipe "includes"
  templates/scripts/            # build.sh, check_updates.sh, fetch_source.sh, make_images.sh, screenshot.sh, update_recipe_list.sh,
                                # check_port.sh (+ port_rules.json), package.sh, test_port.sh, make_controls.sh
  templates/Controls4PM/        # Love2D controls-image creator/viewer (src/converter.lua, src/controls.love)
  templates/PM_MAPPINGS_TEMPLATE.svg  # controls image template used by make_controls.sh
new_ports/<port>/
  source/ , source.tar.gz|zip   # upstream source (gitignored), from check_updates.sh / fetch_source.sh
  <port>/                       # exactly what goes in <port>.zip
    Port Name.sh
    <port>/
      port.json, version.json, gameinfo.xml, README.md, screenshot, cover
      <exe>.gptk, licenses/, libs.<arch>/, conf/
      <exe>.aarch64, <exe>.x86_64   # written by the build workflow
  test/                         # test_port.sh results: report.md, <dev>.png, logs (gitignored)
releases/ports.json             # PortMaster-Info ports.json, used by new_recipe.sh
.github/docker/pm_dockerfile.<arch>   # builder images (aarch64 is Ubuntu 20.04 for glibc compatibility)
```

The port folder name, the zip name (`<port>.zip`), the recipe folder and the `recipe_file` dropdown entry are all the same lowercase string.

## Workflows

- `build_recipe.yml` (manual dispatch, select port): builds the builder image if needed, runs `fetch_source.sh` and then `recipes/ports/<port>/build.sh` per arch, captures a screenshot from the x86_64 build if the recipe has a `screenshot` config, extracts the binaries into the port folder, creates missing `screenshot.png`/`cover.png` (`make_images.sh`), then runs `package.sh`: copies the recipe's `includes` from `recipes/templates/includes/` into the port folder, writes `version.json`, generates the controls image if the recipe has `controls` (`make_controls.sh`), runs `check_port.sh` (errors fail the build) and zips `new_ports/<port>/<port>/` to `<port>.zip` (without `.gitkeep` files), commits the port folder, and publishes the zip as GitHub release `<port>-<version>`.
- `check_recipe_updates.yml` (daily cron): runs `check_updates.sh`, which updates `port_version`/`port_checksum`/`port_download`/`date_updated` in the recipes, refreshes `new_ports/<port>/source` locally, seeds `new_ports/<port>/<port>/` from the published PortMaster zip only if that folder does not exist yet, and rewrites the Port Status table in README.md. Rolling tags such as `continuous` are compared by checksum and stored as `<tag>-<date>`.

Upstream source is not committed. `fetch_source.sh <port>` downloads `port_download`, checks it against `port_checksum` and extracts it to `new_ports/<port>/source`. Branch-based ports are pinned to the commit (`archive/<sha>.zip`), so a checksum mismatch only happens when a tag is moved (e.g. `continuous`). Run Check Recipe Updates, then build. Patches to upstream code belong in the port's `build.sh`, never in `source/`.
- `sync_recipe_list.yml`: runs `update_recipe_list.sh` to regenerate the dropdown between the `recipe-list-start`/`recipe-list-end` markers. Never hand-edit that list.

## Adding or changing a port

1. Run `recipes/new_recipe.sh` from `recipes/`, or copy an existing recipe and port of the same type (native SDL2, Rust/macroquad, Go, etc.). Prefer recently built ports as templates.
2. `recipe.json`: `port_json` must point at `new_ports/<port>/<port>/<port>/port.json`. Leave `port_version`/`port_checksum`/`port_download` blank; `check_updates.sh <port>` fills them in.
3. `build.sh`: keep it to the build command plus whatever is unavoidable. It must leave `dist/<exe>.<arch>` (and any required `dist/libs.<arch>/`) so the tarball extracts into the port folder.
4. `version.json` (`{"name": "<port>.zip", "version": "<port_version>"}`) is generated from the recipe at package time. Don't edit it by hand. To regenerate locally:
   `jq '{name, version: .source.port_version}' recipes/ports/<port>/recipe.json > new_ports/<port>/<port>/<port>/version.json`
5. Run `recipes/templates/scripts/update_recipe_list.sh`, then dispatch the build.

Don't touch other ports' files while working on one port.

## Launch scripts (PortMaster rules)

- Copy the structure of `recipes/templates/port_template/zz_port.sh` or a recent port of the same type. Keep scripts as short as comparable ports; you must be able to explain every non-standard line.
- LF line endings only (`.sh`, `.gptk`). CRLF gives a black screen with no log.
- Use the helpers: `$ESUDO` (not `sudo`), `$controlfolder`, `get_controls`, `$GPTOKEYB`, `$DEVICE_ARCH`, `bind_directories`/`bind_files` for saves (not raw symlinks). End with `pm_finish`, with no `trap` or `pkill gptokeyb`.
- Scripts must work on every CFW (AmberELEC, ArkOS/dArkOS, muOS, ROCKNIX libmali/panfrost, Knulli) and every resolution (640x480 up). No CFW-specific paths.
- Don't set environment variables the default behaviour already covers.
- Code shared by several launchers lives once in `recipes/templates/includes/` and is listed in the recipe (`"includes": ["swap.txt"]`); the launcher does `source "$GAMEDIR/swap.txt"`. Edit the master copy only; the build overwrites the per-port copies. `swap.txt` gives `swap_on`/`swap_off` (temporary zram/swapfile for ROM extraction on <=1GB devices) and is used by lighthouse, soh and starship. `launcher/main.lua`, `osk.lua` and `conf.lua` are a shared Love2D settings launcher (needs the `love_11.5` runtime): each port adds only `launcher/config.lua` (title, items, load/save of the game config; see starship). Anchor co-op ports (lighthouse, soh) also include the shared `launcher/anchor.lua`, and their `config.lua` is one line: `return require("anchor")("Title", { fps = <original fps> })` and runs it from its `.sh`, continuing only if `launch.cfg` sets `LAUNCHER_PLAY=1`. `extract/main.lua` and `conf.lua` (also `love_11.5`) extract bundled archives on first run with a message on screen: `$LOVE_RUN "$GAMEDIR/extract" [--text MSG] [--dir DIR] ARCHIVE... || exit 1` handles `.tar.gz`/`.tgz`, `.tar.xz`, `.tar` and `.zip`, exits 0 when done and otherwise shows the error until a button is pressed (see lix, opengolf). Use it instead of `pm_message` around long work: on sudo CFWs (ArkOS/dArkOS) pm_message's window is never closed and stays on screen over the game.
- X11/OpenGL-only games run under the Westonpack runtime (see `new_ports/recwars`). Runtimes listed in `attr.runtime` must exactly match keys in PortMaster's `runtimes.json`.

## Libraries and binaries

- Never bundle or static-link SDL2, libc/libstdc++/libpthread/librt, libGL/libEGL/mesa or GPU drivers. CFWs ship their own. `check_port.sh` takes its lists from `recipes/templates/scripts/port_rules.json` (genres, port.json attrs, arches, banned/common/CFW-provided libs); edit that file, not the script.
- Bundle only libraries that `ldd` proves are missing on target CFWs, in `libs.<arch>/`.
- The aarch64 build targets an old glibc on purpose (Ubuntu 20.04 container). `attr.min_glibc` must be the highest `GLIBC_x.y` the aarch64 binary and `libs.aarch64/` ask for (blank if 2.17), even below 2.31 (e.g. `pow`/`exp` need 2.29); PortMaster-New flags a blank value. `package.sh` sets it on every build (`check_port.sh <port> --fix` does the same locally).
- Zip extraction drops the exec bit, so the launcher should restore it on the binary at runtime. Commit `.sh` files as 644.

## Metadata

- Don't guess `port.json` syntax. Copy it from existing ports or PortMaster-GUI. `name` is `<port>.zip`, `items` is `[ "<Port Name>.sh", "<port>" ]`.
- `gameinfo.xml` must match the files on disk: root `<gameList>`, and each `<game>` needs non-empty `path`, `name`, `desc`, `releasedate` (`YYYYMMDDT000000`), `developer`, `publisher`, `genre` and `image` (`./<port>/cover.png`). Copy the layout from starship.
- Use one flat `licenses/` folder with one `LICENSE.<component>.txt` per bundled component (game, libs, fonts, assets).
- `screenshot.png` must be 640x480 and both it and `cover.png` under 100 kB; `package.sh` (`check_port.sh --fix`) letterboxes the screenshot and re-encodes oversized images with fewer palette colours (pngquant, else ImageMagick). To have the build capture it, add a `screenshot` object to `recipe.json`; `screenshot.sh` runs the x86_64 binary under Xvfb in the builder image, presses the given keys and saves `screenshot.png` (replacing the existing one on every build). Example from frozenbubble: `"screenshot": {"delay": 8, "keys": "Return Return", "env": "FB_DATA_DIR=./share"}`. `new_recipe.sh` adds `"screenshot": {"delay": 8}` (title-screen capture) only for ports not already in PortMaster, which keep their published screenshot; tune it after the first build, and remove it for ports that need user-supplied game data (soh, starship, lighthouse) or have a hand-taken screenshot. If a port has no `screenshot.png`/`cover.png`, the build copies one to the other or generates a title-card placeholder (PNG comment `PortMaster placeholder`, flagged by `check_port.sh`) and logs a warning. Replace placeholders with a real screenshot before submitting: `test_port.sh <port> --screenshot DEV` takes it on a device (use a 4:3 640x480 one) and removes the recipe's `screenshot` config. `screenshot.sh` unpacks a port's `assets.tar.gz` before running the binary. That also works for GLES ports and ports that need game data, which Xvfb can't capture. No `.gitkeep` files or nested duplicate folders in the final zip.
- Controls image: add `"controls": "launcher/controls.png"` to `recipe.json` and a `| Button | Action |` table to the port's `README.md`. `make_controls.sh` (run by `package.sh`; needs `lua` and `inkscape`, which CI installs only for such recipes) fills `PM_MAPPINGS_TEMPLATE.svg` from the table, the `.gptk` and the port.json title with Controls4PM's converter and writes the PNG to that path. In ports with the shared settings launcher, `launcher/controls.png` adds a Controls item to the menu. Other ports can use `"controls/controls.png"` and show it with `controls.love` before the game (see edgar). The launcher's `conf.lua` sets `LAUNCHER_AUTOSTART` (seconds before Play is chosen automatically, cancelled by any button; 0 = off).
- Keep the original porter's credit when building on someone else's port. No DRM bypass or Steam emulators.

## Testing before submitting to PortMaster-New

Test on several CFWs and SoCs (rk3326, rk3566, h700, a133p, sm8250/Adreno), not just one device. A port tested on a single device gets rejected.

1. `recipes/templates/scripts/check_port.sh <port>`: static checks (CRLF, launcher rules, port.json items/runtimes/`min_glibc`, forbidden bundled libs, gameinfo.xml, licenses, placeholders) and writes the result to `new_ports/<port>/check.md`. The build runs it with `--fix --fix-bloat --fix-license`; `--fix` sets `min_glibc`, shrinks/letterboxes the images, sets the launcher to 644, fixes SC2188, drops built-in runtimes from `attr.runtime` and renames `LICENSE_x` files; it fails on the errors left. `--test` changes nothing in the port and lists (on screen and in `check.md`) what the three fix flags would change plus the errors/warnings left for a manual fix. `--fix-bloat` deletes bundled `common_libs` (also drop them from `FILES` in `build.sh`). Every bundled lib needs a matching `licenses/LICENSE.<component>.txt` (warning; `license_aliases` in `port_rules.json` maps e.g. libcrypto to openssl); `--fix-license` downloads missing ones from `license_urls` there.
2. `recipes/templates/scripts/test_port.sh <port> [DEV...]` (local; needs DeviceHub's `devhub`, or `DEVHUB=/path/to/devhub`): packages the port with `package.sh` (or `--zip FILE|URL`), then on every online device (or the ones given) installs it, starts the launcher with the front end stopped, records FPS, presses `test.keys`, takes a screenshot, runs `ldd` on the binaries, pulls `log.txt`, and leaves the device clean and released. PASS = launcher still running, screen not blank, no missing libs. Results go to `new_ports/<port>/test/` (gitignored).
3. Paste `new_ports/<port>/test/report.md` (one row per device: CFW, SoC/GPU, resolution, result, FPS) into the PortMaster-New PR.

Optional `"test"` object in `recipe.json`: `{"delay": 15, "keys": "a 3 start", "env": "FOO=1", "needs_data": "oot.o2r baseroms/*.z64"}`. `keys` are `devhub input` tokens (`a b x y l1 r1 start select up down left right key:enter`, a number waits) sent to a virtual gamepad, so they go through gptokeyb and the `.gptk`. Button names are positional (Xbox layout, `a` = bottom button); Nintendo-layout devices may swap A/B and X/Y, so prefer `start`, the D-pad and `key:` tokens. `needs_data` lists user-supplied files (globs in the port folder); when they are not on the device only the launcher's error path is tested. Use one feature branch per port in PortMaster-New and push fixes to the existing PR instead of resubmitting.
