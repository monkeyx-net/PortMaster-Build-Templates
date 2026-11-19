-- Frame rate and N64 low-res mode (shared logic in anchor.lua; Ghostship has no Anchor co-op);
-- also keeps the menu bar closed so the game starts straight away. Defaults to the original 30 fps.
-- Console logging defaults to warnings and errors (LogLevel 3) instead of info.
return require("anchor")("Ghostship", { fps = 30, anchor = false, settings = { OpenMenuBar = 0 }, custom = { 720, 600 },
    defaults = { gDeveloperTools = { LogLevel = 3 } } })
