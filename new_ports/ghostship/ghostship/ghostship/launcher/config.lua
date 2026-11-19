-- Frame rate and N64 low-res mode (shared logic in anchor.lua; Ghostship has no Anchor co-op);
-- also keeps the menu bar closed so the game starts straight away. Defaults to the original 30 fps.
-- Console logging defaults to warnings and errors (LogLevel 3) instead of info.
-- Texture filter (gSettings.TextureFilter: 0 three-point, 1 linear, 2 none) defaults to linear.
return require("anchor")("Ghostship", { fps = 30, anchor = false, settings = { OpenMenuBar = 0 }, custom = { 720, 600 },
    defaults = { gDeveloperTools = { LogLevel = 3 } },
    items = { { id = "filter", label = "Texture filter", kind = "choice",
                choices = { { 1, "Linear" }, { 0, "Three-point" }, { 2, "None (pixelated)" } },
                help = "Linear is smooth and fast. Three-point looks like the N64 but costs more GPU." } },
    keys = { filter = { { "gSettings", "TextureFilter" }, 1 } } })
