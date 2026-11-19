-- Frame rate and N64 low-res mode (shared logic in anchor.lua; Ghostship has no Anchor co-op);
-- also keeps the menu bar closed so the game starts straight away. Defaults to the original 30 fps.
-- Texture filter (gSettings.TextureFilter: 0 three-point, 1 linear, 2 none) defaults to linear.
-- Log level (gDeveloperTools.LogLevel, spdlog levels) defaults to warnings instead of info.
return require("anchor")("Ghostship", { fps = 30, anchor = false, settings = { OpenMenuBar = 0 }, custom = { 720, 600 },
    items = {
        { id = "filter", label = "Texture filter", kind = "choice",
          choices = { { 1, "Linear" }, { 0, "Three-point" }, { 2, "None (pixelated)" } },
          help = "Linear is smooth and fast. Three-point looks like the N64 but costs more GPU." },
        { id = "loglevel", label = "Log level", kind = "choice",
          choices = { { 3, "Warnings" }, { 4, "Errors" }, { 2, "Info" }, { 1, "Debug" }, { 6, "Off" } },
          help = "How much the game writes to log.txt. Info or Debug help when reporting a problem." },
    },
    keys = {
        filter = { { "gSettings", "TextureFilter" }, 1 },
        loglevel = { { "gDeveloperTools", "LogLevel" }, 3 },
    } })
