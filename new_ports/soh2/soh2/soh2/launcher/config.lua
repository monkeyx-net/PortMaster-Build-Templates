-- Frame rate and N64 low-res mode (shared logic in anchor.lua; 2Ship has no Anchor co-op and
-- keeps these CVars flat: gInterpolationFPS, gLowResMode, ...)
return require("anchor")("Ship of Harkinian 2", { fps = 20, o2r = { "mm.o2r" }, anchor = false, flat = true })
