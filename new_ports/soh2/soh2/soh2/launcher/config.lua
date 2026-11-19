-- Ship of Harkinian 2 launcher settings (shared logic in cvars.lua; 2Ship has no Anchor co-op);
-- also closes the game menu and forces gSettings.ControlNav to 1. 5.0.1 keeps libultraship's
-- CVars in gSettings (VsyncEnabled, LowResMode, ControlNav, TextureFilter, AdvancedResolution).
return require("cvars")({
    title = "Ship of Harkinian 2",
    subtitle = "Majora's Mask settings",
    help = "In game: press Select to open the menu.",
    items = {
        { id = "fps", label = "Frame rate", kind = "choice",
          choices = { { 20, "20 (original)" }, { 30, "30" }, { 40, "40" }, { 60, "60" } },
          help = "Higher is smoother but needs a faster device. With VSync on it is capped at the display rate." },
        { id = "matchrate", label = "Match display rate", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "On runs at the display's refresh rate and ignores the frame rate above." },
        { id = "vsync", label = "VSync", kind = "choice",
          choices = { { 1, "On" }, { 0, "Off" } },
          help = "On removes tearing but caps the frame rate at the display's refresh rate." },
        { id = "lowres", label = "N64 low-res mode", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "Renders at the N64's original resolution. Helps on weaker devices." },
        "filter",
        "loglevel",
        "resolution",
    },
    keys = {
        fps = { { "gInterpolationFPS" }, 20 },
        matchrate = { { "gMatchRefreshRate" }, 0 },
        vsync = { { "gSettings", "VsyncEnabled" }, 1 },
        lowres = { { "gSettings", "LowResMode" }, 0 },
    },
    forced = { { { "gWindows", "Menu" }, 0 }, { { "gSettings", "ControlNav" }, 1 } },
    resolution = { "gSettings", "AdvancedResolution" },
    filter = { "gSettings", "TextureFilter" },
    o2r = { "mm.o2r" },
})
