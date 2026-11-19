-- SpaghettiKart launcher settings (shared logic in cvars.lua); also forces gOpenMenu to 0
-- (port menu closed) and gControlNav to 1
return require("cvars")({
    title = "SpaghettiKart",
    subtitle = "Mario Kart 64 settings",
    help = "In game: press Select to open the menu.",
    items = {
        { id = "fps", label = "Frame rate", kind = "choice",
          choices = { { 30, "30 (original)" }, { 40, "40" }, { 45, "45" }, { 50, "50" }, { 60, "60" },
                      { 90, "90" }, { 120, "120" } },
          help = "Higher is smoother but needs a faster device. With VSync on it is capped at the display rate." },
        { id = "vsync", label = "VSync", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "On removes tearing but caps the frame rate at the display's refresh rate." },
        { id = "lod", label = "Level of detail", kind = "choice",
          choices = { { 1, "Full detail" }, { 0, "Original (LOD)" } },
          help = "Original draws simpler cars, trucks, trains and plants in the distance. Helps on weaker devices." },
        { id = "culling", label = "Disable culling", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "On also draws objects off screen and far away. Off (original) is much faster." },
        "filter",       -- gTextureFilter
        "loglevel",
        "resolution",
    },
    keys = {
        fps = { { "gInterpolationFPS" }, 30 },
        vsync = { { "gVsyncEnabled" }, 0 },        -- the game's default is 1
        lod = { { "gDisableLod" }, 1 },
        culling = { { "gNoCulling" }, 0 },
    },
    forced = { { { "gOpenMenu" }, 0 }, { { "gControlNav" }, 1 } },
})
