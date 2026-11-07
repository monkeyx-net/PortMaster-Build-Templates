-- Pikifen launcher settings, saved as shell variables in pikifen.cfg (sourced by Pikifen.sh)

local config = {
    title = "Pikifen",
    subtitle = "Pikmin fan game",
    help = "Left stick moves, right stick aims, A throws.",
    items = {
        { id = "MOUSE_SCALE", label = "Cursor speed", kind = "choice",
          choices = { { 8192, "Slow" }, { 6144, "Normal" }, { 4096, "Fast" }, { 3072, "Faster" } },
          help = "How fast the right stick moves the cursor." },
        { id = "SCREEN_MODE", label = "Fullscreen", kind = "choice",
          choices = { { 1, "Hardware" }, { 0, "Software" } },
          help = "Hardware runs at the device resolution. Try Software if the picture is wrong." },
        { id = "FPS", label = "Frame rate", kind = "choice",
          choices = { { 60, "60" }, { 30, "30" } },
          help = "Use 30 on slower devices." },
    },
}

local DEFAULTS = { MOUSE_SCALE = 6144, SCREEN_MODE = 1, FPS = 60 }

function config.load(path)
    local v = {}
    for k, d in pairs(DEFAULTS) do v[k] = d end
    local f = io.open(path, "r")
    if f then
        for l in f:lines() do
            local k, n = l:match("^([%w_]+)=(%d+)$")
            if k and v[k] then v[k] = tonumber(n) end
        end
        f:close()
    end
    return v
end

function config.save(path, v)
    local f, err = io.open(path, "w")
    if not f then return false, err end
    for _, it in ipairs(config.items) do
        f:write(it.id, "=", v[it.id], "\n")
    end
    f:close()
    return true
end

return config
