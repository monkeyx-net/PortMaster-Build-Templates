-- Starship launcher settings. Edits top-level CVars in starship.cfg.json line by
-- line; the game writes it pretty-printed (one key per line), so no JSON parser
-- is needed. Also forces gOpenMenuBar to 0 (menu closed) and gControlNav to 1.

local config = {
    title = "Starship",
    subtitle = "Star Fox 64 settings",
    help = "In game: press Select, then X to access the menu bar.",
    items = {
        { id = "gInterpolationFPS", label = "Frame rate", kind = "choice",
          choices = { { 30, "30 (original)" }, { 60, "60" } },
          help = "60 is smoother but needs a faster device." },
        { id = "gLowResMode", label = "N64 low-res mode", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } },
          help = "Renders at the N64's original resolution. Helps on weaker devices." },
        { id = "gVoiceLanguage", label = "Voices", kind = "choice",
          choices = { { 0, "Original" }, { 1, "Lylat (EU)" } },
          help = "Lylat voices need an EU Star Fox 64 rom in baseroms/ when the game data is first built." },
        { id = "gInvertYAxis", label = "Invert Y axis", kind = "choice",
          choices = { { 0, "Off" }, { 1, "On" } } },
    },
}

-- Game defaults when a key is not in the file yet
local DEFAULTS = { gInterpolationFPS = 60, gLowResMode = 0, gVoiceLanguage = 0, gInvertYAxis = 0 }
-- Always written on save
local FORCED = { gOpenMenuBar = 0, gControlNav = 1 }

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
    -- values the launcher doesn't offer (e.g. 20 fps) snap to the nearest choice
    for _, it in ipairs(config.items) do
        local ok = false
        for _, c in ipairs(it.choices) do if c[1] == v[it.id] then ok = true end end
        if not ok then v[it.id] = (it.id == "gInterpolationFPS" and v[it.id] > 30) and 60 or it.choices[1][1] end
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
