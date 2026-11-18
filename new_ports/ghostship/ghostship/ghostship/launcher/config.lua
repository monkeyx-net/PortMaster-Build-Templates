-- Frame rate and N64 low-res mode (shared logic in anchor.lua; Ghostship has no Anchor co-op);
-- also keeps the menu bar closed so the game starts straight away. Unset, the game runs at 60 fps.
return require("anchor")("Ghostship", { fps = 30, fpsDefault = 60, anchor = false, settings = { OpenMenuBar = 0 } })
