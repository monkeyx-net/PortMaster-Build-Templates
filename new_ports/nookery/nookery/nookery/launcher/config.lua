-- Nookery launcher settings. The frame rate is "video.maxFrameRate" in the game's config.json,
-- which the game writes one key per line, so it is edited line by line. Aspect and render
-- resolution are port options (patched into the game): they are saved to launcher.cfg next to
-- config.json, which Nookery.sh sources and passes on as NOOKERY_ASPECT / NOOKERY_RENDER_SCALE.

local config = {
    title = "Animal Crossing",
    subtitle = "Nookery settings",
    help = "Start plays the game. Settings are saved when the game starts.",
    items = {
        { id = "fps", label = "Frame rate", kind = "choice",
          choices = { { 30, "30" }, { 60, "60" } },
          help = "30 halves the work on slower devices." },
        { id = "aspect", label = "Aspect", kind = "choice",
          choices = { { "auto", "Auto" }, { "4:3", "4:3" } },
          help = "Auto widens the view on wide screens and letterboxes on square ones." },
        { id = "scale", label = "Render resolution", kind = "choice",
          choices = { { "0", "Screen" }, { "1", "640x480 (GameCube)" } },
          help = "640x480 is faster on high-resolution screens." },
    },
}

local DEFAULTS = { fps = 60, aspect = "auto", scale = "0" }

local function readLines(path)
    local f = io.open(path, "r")
    if not f then return nil end
    local lines = {}
    for l in f:lines() do lines[#lines + 1] = l end
    f:close()
    return lines
end

local function writeFile(path, text)
    local tmp = path .. ".tmp"
    local f, err = io.open(tmp, "w")
    if not f then return false, err end
    f:write(text)
    f:close()
    return os.rename(tmp, path)
end

local function extrasPath(path)
    return (path:match("^(.*)/") or ".") .. "/launcher.cfg"
end

function config.load(path)
    local v = {}
    for k, d in pairs(DEFAULTS) do v[k] = d end
    for _, l in ipairs(readLines(path) or {}) do
        local n = tonumber(l:match('"video%.maxFrameRate":%s*(%d+)'))
        if n then v.fps = n end
    end
    for _, l in ipairs(readLines(extrasPath(path)) or {}) do
        v.aspect = l:match("^NOOKERY_ASPECT=(%S+)") or v.aspect
        v.scale = l:match("^NOOKERY_RENDER_SCALE=(%S+)") or v.scale
    end
    -- keep a frame rate set in game (e.g. 0 = unlimited) as an extra choice
    local ok = false
    for _, c in ipairs(config.items[1].choices) do if c[1] == v.fps then ok = true end end
    if not ok then table.insert(config.items[1].choices, { v.fps, tostring(v.fps) .. " (set in game)" }) end
    return v
end

function config.save(path, v)
    local lines = readLines(path)
    if not lines then return false, "cannot read " .. path end
    local found = false
    for i, l in ipairs(lines) do
        if l:match('"video%.maxFrameRate":') then
            lines[i] = l:gsub('("video%.maxFrameRate":%s*)%d+', "%1" .. v.fps)
            found = true
        end
    end
    if not found then
        -- add it as the first key; "$version" always follows, so it needs a comma
        for i, l in ipairs(lines) do
            if l:match("^%s*{%s*$") then
                table.insert(lines, i + 1, '    "video.maxFrameRate": ' .. v.fps .. ",")
                break
            end
        end
    end
    local ok, err = writeFile(path, table.concat(lines, "\n") .. "\n")
    if not ok then return false, err end
    return writeFile(extrasPath(path),
        "NOOKERY_ASPECT=" .. v.aspect .. "\nNOOKERY_RENDER_SCALE=" .. v.scale .. "\n")
end

return config
