-- Lix launcher settings, saved as shell variables in lix.cfg (sourced by Lix.sh)

local config = {
    title = "Lix",
    subtitle = "Lemmings-like puzzle game",
    help = "Left stick moves the mouse, A clicks.",
    items = {
        { id = "LIX_CURSOR_MIN_SCALE", label = "Mouse cursor size", kind = "choice",
          choices = { { 1, "Normal" }, { 2, "2x" }, { 3, "3x" }, { 4, "4x" } },
          help = "Minimum cursor scale. Bigger is easier to see on small screens." },
    },
}

local DEFAULTS = { LIX_CURSOR_MIN_SCALE = 2 }

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
    for k in pairs(DEFAULTS) do f:write(k, "=", v[k], "\n") end
    f:close()
    return true
end

return config
