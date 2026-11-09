-- SpaghettiKart launcher settings. Edits top-level CVars in spaghettify.cfg.json line by
-- line; the game writes it pretty-printed (one key per line), so no JSON parser
-- is needed. Also forces gOpenMenu to 0 (port menu closed) and gControlNav to 1.

local config = {
    title = "SpaghettiKart",
    subtitle = "Mario Kart 64 settings",
    help = "In game: press Select to open the menu.",
    items = {
        { id = "gInterpolationFPS", label = "Frame rate", kind = "choice",
          choices = { { 30, "30 (original)" }, { 40, "40" }, { 45, "45" }, { 50, "50" }, { 60, "60" },
                      { 90, "90" }, { 120, "120" } },
          help = "Higher is smoother but needs a faster device. With VSync on it is capped at the display rate." },
        { id = "gVsyncEnabled", label = "VSync", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "On removes tearing but caps the frame rate at the display's refresh rate." },
        { id = "gDisableLod", label = "Level of detail", kind = "choice",
          choices = { { 1, "Full detail" }, { 0, "Original (LOD)" } },
          help = "Original draws simpler cars, trucks, trains and plants in the distance. Helps on weaker devices." },
        { id = "gNoCulling", label = "Disable culling", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "On also draws objects off screen and far away. Off (original) is much faster." },
    },
}

-- Game defaults when a key is not in the file yet (VSync is on in the game, off here)
local DEFAULTS = { gInterpolationFPS = 30, gVsyncEnabled = 0, gDisableLod = 1, gNoCulling = 0 }
-- Always written on save
local FORCED = { gOpenMenu = 0, gControlNav = 1 }

local function readLines(path)
    local f = io.open(path, "r")
    if not f then return nil end
    local lines = {}
    for l in f:lines() do lines[#lines + 1] = l end
    f:close()
    return lines
end

-- Set a numeric CVar, inserting it at the top of "CVars" if missing
local function setInt(lines, key, value)
    for i, l in ipairs(lines) do
        if l:match('"' .. key .. '":') then
            lines[i] = l:gsub('("' .. key .. '":%s*)[%d%.%-]+', '%1' .. value)
            return
        end
    end
    for i, l in ipairs(lines) do
        local indent = l:match('^(%s*)"CVars":%s*{%s*$')
        if indent then
            local last = (lines[i + 1] or ""):match("^%s*}")
            table.insert(lines, i + 1, indent .. '    "' .. key .. '": ' .. value .. (last and "" or ","))
            return
        end
    end
end

function config.load(path)
    local v = {}
    for k, d in pairs(DEFAULTS) do v[k] = d end
    for _, l in ipairs(readLines(path) or {}) do
        for k in pairs(DEFAULTS) do
            local n = tonumber(l:match('"' .. k .. '":%s*([%d%.%-]+)'))
            if n then v[k] = n end
        end
    end
    -- a value set in game that the launcher doesn't offer (e.g. 45 fps) is kept as an extra choice
    for _, it in ipairs(config.items) do
        local ok = false
        for _, c in ipairs(it.choices) do if c[1] == v[it.id] then ok = true end end
        if not ok then table.insert(it.choices, { v[it.id], tostring(v[it.id]) .. " (set in game)" }) end
    end
    return v
end

function config.save(path, v)
    local lines = readLines(path)
    if not lines then return false, "cannot read " .. path end
    for k in pairs(DEFAULTS) do setInt(lines, k, v[k]) end
    for k, n in pairs(FORCED) do setInt(lines, k, n) end
    local tmp = path .. ".tmp"
    local f, err = io.open(tmp, "w")
    if not f then return false, err end
    f:write(table.concat(lines, "\n"), "\n")
    f:close()
    return os.rename(tmp, path)
end

return config
