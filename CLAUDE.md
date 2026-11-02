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
  templates/scripts/            # build.sh, check_updates.sh, fetch_source.sh, make_images.sh, screenshot.sh, update_recipe_list.sh
new_ports/<port>/
  source/ , source.tar.gz|zip   # upstream source (gitignored), from check_updates.sh / fetch_source.sh
  <port>/                       # exactly what goes in <port>.zip
    Port Name.sh
    <port>/
      port.json, version.json, gameinfo.xml, README.md, screenshot, cover
      <exe>.gptk, licenses/, libs.<arch>/, conf/
      <exe>.aarch64, <exe>.x86_64   # written by the build workflow
releases/ports.json             # PortMaster-Info ports.json, used by new_recipe.sh
.github/docker/pm_dockerfile.<arch>   # builder images (aarch64 is Ubuntu 20.04 for glibc compatibility)
```

The port folder name, the zip name (`<port>.zip`), the recipe folder and the `recipe_file` dropdown entry are all the same lowercase string.

## Workflows

- `build_recipe.yml` (manual dispatch, select port): builds the builder image if needed, runs `fetch_source.sh` and then `recipes/ports/<port>/build.sh` per arch, captures a screenshot from the x86_64 build if the recipe has a `screenshot` config, extracts the binaries into the port folder, copies the recipe's `includes` from `recipes/templates/includes/` into the port folder, creates missing `screenshot.png`/`cover.png` (`make_images.sh`), writes `version.json`, zips `new_ports/<port>/<port>/` to `<port>.zip` (without `.gitkeep` files), commits the port folder, and publishes the zip as GitHub release `<port>-<version>`.
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
- Code shared by several launchers lives once in `recipes/templates/includes/` and is listed in the recipe (`"includes": ["swap.txt"]`); the launcher does `source "$GAMEDIR/swap.txt"`. Edit the master copy only; the build overwrites the per-port copies. `swap.txt` gives `swap_on`/`swap_off` (temporary zram/swapfile for ROM extraction on <=1GB devices) and is used by lighthouse, soh and starship. `launcher/main.lua`, `osk.lua` and `conf.lua` are a shared Love2D settings launcher (needs the `love_11.5` runtime): each port adds only `launcher/config.lua` (title, items, load/save of the game config; see starship). Anchor co-op ports (lighthouse, soh) also include the shared `launcher/anchor.lua`, and their `config.lua` is one line: `return require("anchor")("Title")` and runs it from its `.sh`, continuing only if `launch.cfg` sets `LAUNCHER_PLAY=1`.
- X11/OpenGL-only games run under the Westonpack runtime (see `new_ports/recwars`). Runtimes listed in `attr.runtime` must exactly match keys in PortMaster's `runtimes.json`.

## Libraries and binaries

- Never bundle or static-link SDL2, libc/libstdc++/libpthread/librt, libGL/libEGL/mesa or GPU drivers. CFWs ship their own.
- Bundle only libraries that `ldd` proves are missing on target CFWs, in `libs.<arch>/`.
- The aarch64 build targets an old glibc on purpose (Ubuntu 20.04 container). Keep `min_glibc` in port.json accurate.
- Zip extraction drops the exec bit, so the launcher should restore it on the binary at runtime. Commit `.sh` files as 644.

## Metadata

- Don't guess `port.json` syntax. Copy it from existing ports or PortMaster-GUI. `name` is `<port>.zip`, `items` is `[ "<Port Name>.sh", "<port>" ]`.
- `gameinfo.xml` must match the files on disk.
- Use one flat `licenses/` folder with one `LICENSE.<component>.txt` per bundled component (game, libs, fonts, assets).
- Screenshot is 640x480. To have the build capture it, add a `screenshot` object to `recipe.json`; `screenshot.sh` runs the x86_64 binary under Xvfb in the builder image, presses the given keys and saves `screenshot.png` (replacing the existing one on every build). Example from frozenbubble: `"screenshot": {"delay": 8, "keys": "Return Return", "env": "FB_DATA_DIR=./share"}`. `new_recipe.sh` adds `"screenshot": {"delay": 8}` (title-screen capture) only for ports not already in PortMaster, which keep their published screenshot; tune it after the first build, and remove it for ports that need user-supplied game data (soh, starship, lighthouse) or have a hand-taken screenshot. If a port has no `screenshot.png`/`cover.png`, the build copies one to the other or generates a title-card placeholder and logs a warning. Replace placeholders with a real screenshot before submitting. No `.gitkeep` files or nested duplicate folders in the final zip.
- Keep the original porter's credit when building on someone else's port. No DRM bypass or Steam emulators.

## Testing before submitting to PortMaster-New

Test on several CFWs and SoCs (rk3326, rk3566, h700, a133p, sm8250/Adreno), not just one device. A port tested on a single device gets rejected. Use one feature branch per port in PortMaster-New and push fixes to the existing PR instead of resubmitting.
