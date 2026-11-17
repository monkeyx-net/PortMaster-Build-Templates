-- Frame rate and N64 low-res mode (shared logic in anchor.lua; PaperBoat has no Anchor co-op);
-- also keeps the menu bar closed so the game starts straight away
return require("anchor")("PaperBoat", { fps = 30, o2r = { "pm64.o2r" }, anchor = false, settings = { OpenMenuBar = 0 } })
