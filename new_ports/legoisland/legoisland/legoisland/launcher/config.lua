-- LEGO Island launcher settings. Edits isle.ini (iniparser format: [section], then
-- lowercase `key = "value"` lines) and always points the save path at the port folder.

-- "3D Device ID": driver number and the miniwin renderer GUID as printed by the game
local function device(n) return "0 0x682656f3 0x0 0x0 0x" .. n .. "000000" end

local config = {
    title = "LEGO Island",
    subtitle = "isle-portable settings",
    help = "In game: D-pad moves, right or left stick moves the cursor, A clicks.",
    items = {
        { id = "isle:3d device id", label = "Renderer", kind = "choice",
          choices = { { "", "Automatic" }, { device(4), "OpenGL ES 3" }, { device(6), "OpenGL ES 2" },
                      { device(2), "Software" } },
          help = "Try OpenGL ES 2 or Software if the picture is wrong." },
        { id = "isle:island quality", label = "Island detail", kind = "choice",
          choices = { { "0", "Low" }, { "1", "Medium" }, { "2", "High" } },
          help = "Lower it on slower devices." },
        { id = "isle:island texture", label = "Island textures", kind = "choice",
          choices = { { "0", "Low" }, { "1", "High" } } },
        { id = "isle:frame delta", label = "Frame rate limit", kind = "choice",
          choices = { { "10", "100" }, { "16", "60" }, { "33", "30" } } },
        { id = "isle:transition type", label = "Transitions", kind = "choice",
          choices = { { "1", "None" }, { "2", "Dissolve" }, { "3", "Mosaic" }, { "4", "Wipe down" },
                      { "5", "Windows" } } },
        { id = "isle:cursor sensitivity", label = "Cursor speed", kind = "choice",
          choices = { { "2.000000", "Slow" }, { "4.000000", "Normal" }, { "6.000000", "Fast" } } },
        { id = "isle:music", label = "Music", kind = "choice",
          choices = { { "true", "On" }, { "false", "Off" } } },
        { id = "extensions:third person camera", label = "Third-person camera", kind = "choice",
          choices = { { "false", "Off" }, { "true", "On" } },
          help = "isle-portable extension: follow the player from behind." },
    },
}

local function readLines(path)
    local f = io.open(path, "r")
    if not f then return nil end
    local lines = {}
    for l in f:lines() do lines[#lines + 1] = l end
    f:close()
    return lines
end

-- Calls fn(section, key, index) for every key line
local function eachKey(lines, fn)
    local section
    for i, l in ipairs(lines) do
        section = l:match("^%s*%[(.-)%]") or section
        local key = l:match("^%s*([^=%[;#]-)%s*=")
        if section and key then fn(section:lower(), key:lower(), i) end
    end
end

local function setValue(lines, id, value)
    local section, key = id:match("^(.-):(.*)$")
    local line = string.format('%-30s = "%s"', key, value)
    local found, header
    eachKey(lines, function(s, k, i) if s == section and k == key then found = i end end)
    if found then lines[found] = line return end
    for i, l in ipairs(lines) do
        if (l:match("^%s*%[(.-)%]") or ""):lower() == section then header = i end
    end
    if not header then
        lines[#lines + 1] = "[" .. section .. "]"
        header = #lines
    end
    table.insert(lines, header + 1, line)
end

local function removeValue(lines, id)
    local found
    eachKey(lines, function(s, k, i) if s .. ":" .. k == id then found = i end end)
    if found then table.remove(lines, found) end
end

function config.load(path)
    local v = {}
    for _, it in ipairs(config.items) do v[it.id] = it.choices[1][1] end
    local lines = readLines(path) or {}
    eachKey(lines, function(s, k, i)
        local id = s .. ":" .. k
        if v[id] ~= nil then v[id] = lines[i]:match('=%s*"?(.-)"?%s*$') end
    end)
    -- a value the launcher doesn't offer is kept as an extra choice
    for _, it in ipairs(config.items) do
        local ok = false
        for _, c in ipairs(it.choices) do if c[1] == v[it.id] then ok = true end end
        if not ok then table.insert(it.choices, { v[it.id], v[it.id] .. " (custom)" }) end
    end
    return v
end

function config.save(path, v)
    local lines = readLines(path)
    if not lines then return false, "cannot read " .. path end
    for _, it in ipairs(config.items) do
        if v[it.id] == "" then removeValue(lines, it.id) else setValue(lines, it.id, v[it.id]) end
    end
    -- The port folder differs per CFW (the .sh sets the game file paths)
    setValue(lines, "isle:savepath", path:match("^(.*)/[^/]*$") .. "/saves/")
    setValue(lines, "isle:full screen", "true")
    local tmp = path .. ".tmp"
    local f, err = io.open(tmp, "w")
    if not f then return false, err end
    f:write(table.concat(lines, "\n"), "\n")
    f:close()
    return os.rename(tmp, path)
end

return config
