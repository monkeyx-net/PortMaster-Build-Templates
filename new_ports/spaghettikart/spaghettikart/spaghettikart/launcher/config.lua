-- SpaghettiKart launcher settings. Edits CVars in spaghettify.cfg.json line by line; the
-- game writes it pretty-printed (one key per line), so no JSON parser is needed. Also
-- forces gOpenMenu to 0 (port menu closed) and gControlNav to 1.

-- Aspect ratio presets as in the game's Graphics menu (ResolutionEditor.cpp); the index is
-- gAdvancedResolution.UIComboItem.AspectRatio. "Off" sets X/Y to 0 (fill the screen).
local ASPECTS = {
    [0] = { 0, 0 }, [1] = { 720, 720 }, [2] = { 4, 3 }, [3] = { 16, 9 }, [4] = { 5, 3 },
    [5] = { 16, 10 }, [6] = { 21, 9 },
}

local config = {
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
        { id = "advres", label = "Advanced resolution", kind = "choice",
          choices = { { 1, "On" }, { 0, "Off" } },
          help = "Enables the aspect ratio setting below (game menu: Graphics, Enable advanced settings)." },
        { id = "aspect", label = "Aspect ratio", kind = "choice",
          choices = { { 0, "Off" }, { 1, "Custom" }, { 2, "Original (4:3)" }, { 3, "Widescreen (16:9)" },
                      { 4, "Nintendo 3DS (5:3)" }, { 5, "16:10 (8:5)" }, { 6, "Ultrawide (21:9)" } },
          help = "Off fills the screen. Custom is 720 x 720 (square) unless changed in the game menu." },
    },
}

-- id -> CVar path below "CVars", and the value used when it's not in the file yet
local KEYS = {
    fps = { { "gInterpolationFPS" }, 30 },
    vsync = { { "gVsyncEnabled" }, 0 },        -- the game's default is 1
    lod = { { "gDisableLod" }, 1 },
    culling = { { "gNoCulling" }, 0 },
    advres = { { "gAdvancedResolution", "Enabled" }, 1 },
    aspect = { { "gAdvancedResolution", "UIComboItem", "AspectRatio" }, 0 },
}
local ASPECT_X = { "gAdvancedResolution", "AspectRatioX" }
local ASPECT_Y = { "gAdvancedResolution", "AspectRatioY" }
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

local function braces(l)
    local _, open = l:gsub("{", "")
    local _, close = l:gsub("}", "")
    return open - close
end

-- Line of the direct child "key" of the object opened on line s (nil if missing), and the
-- line closing that object
local function child(lines, s, key)
    local depth, found = 0, nil
    for i = s + 1, #lines do
        if depth == 0 then
            if lines[i]:match("^%s*}") then return found, i end
            if lines[i]:match('^%s*"' .. key .. '":') then found = found or i end
        end
        depth = depth + braces(lines[i])
    end
end

-- Insert "key": text as the first child of the object opened on line s
local function insert(lines, s, key, text)
    local indent = lines[s]:match("^(%s*)") .. "    "
    local empty = (lines[s + 1] or ""):match("^%s*}")
    table.insert(lines, s + 1, indent .. '"' .. key .. '": ' .. text .. (empty and "" or ","))
end

local function cvarsLine(lines)
    for i, l in ipairs(lines) do if l:match('^%s*"CVars":%s*{%s*$') then return i end end
end

-- Line of a CVar path; with create, missing objects are added and a missing value is
-- inserted as "0"
local function locate(lines, path, create)
    local s = cvarsLine(lines)
    if not s then return nil end
    for n, key in ipairs(path) do
        local i = child(lines, s, key)
        if not i then
            if not create then return nil end
            if n < #path then
                insert(lines, s, key, "{")
                table.insert(lines, s + 2, lines[s + 1]:match("^(%s*)") .. "}" .. (lines[s + 1]:match(",$") and "," or ""))
                lines[s + 1] = lines[s + 1]:gsub(",$", "")
            else
                insert(lines, s, key, "0")
            end
            i = s + 1
        end
        s = i
    end
    return s
end

local function get(lines, path)
    local i = locate(lines, path)
    return i and tonumber(lines[i]:match(':%s*([%d%.%-]+)'))
end

-- Floats keep a decimal point: the game reads 720 as an integer CVar, and CVarGetFloat
-- ignores integer CVars
local function set(lines, path, value, float)
    local i = locate(lines, path, true)
    if not i then return end
    local text = float and string.format("%.1f", value) or tostring(math.floor(value))
    lines[i] = lines[i]:gsub('(":%s*)[%d%.%-]+', "%1" .. text, 1)
end

function config.load(path)
    local lines = readLines(path) or {}
    local v = {}
    for id, k in pairs(KEYS) do v[id] = get(lines, k[1]) or k[2] end
    -- keep a custom ratio set in the game menu
    if v.aspect == 1 then config.customX, config.customY = get(lines, ASPECT_X), get(lines, ASPECT_Y) end
    -- a value set in game that the launcher doesn't offer (e.g. 75 fps) is kept as an extra choice
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
    for id, k in pairs(KEYS) do set(lines, k[1], v[id]) end
    local xy = ASPECTS[v.aspect]
    if v.aspect == 1 and (config.customX or 0) > 0 and (config.customY or 0) > 0 then
        xy = { config.customX, config.customY }
    end
    if xy then
        set(lines, ASPECT_X, xy[1], true)
        set(lines, ASPECT_Y, xy[2], true)
    end
    for k, n in pairs(FORCED) do set(lines, { k }, n) end
    local tmp = path .. ".tmp"
    local f, err = io.open(tmp, "w")
    if not f then return false, err end
    f:write(table.concat(lines, "\n"), "\n")
    f:close()
    return os.rename(tmp, path)
end

return config
