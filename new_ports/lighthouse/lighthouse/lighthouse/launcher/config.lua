-- Reads and edits lighthouse.cfg.json line by line. The game writes it
-- pretty-printed (one key per line), so no JSON parser is needed.
--   Anchor Name / RoomId  - online co-op player name and room
--   gOpenWindows.Menu     - forced to 0 so the game starts with the menu closed
--   gSettings.ControlNav  - forced to 1 so menus work with a controller

local config = {}

local function readLines(path)
    local f = io.open(path, "r")
    if not f then return nil end
    local lines = {}
    for l in f:lines() do lines[#lines + 1] = l end
    f:close()
    return lines
end

-- Calls fn(line, inAnchor) for every line and stores the returned line.
-- inAnchor is true only for keys directly inside the "Anchor" object, so the
-- renderer's "Name" under "Window" is never touched.
local function eachLine(lines, fn)
    local depth = 0
    for i, l in ipairs(lines) do
        lines[i] = fn(l, depth == 1)
        if depth > 0 then
            local _, open = l:gsub("{", "")
            local _, close = l:gsub("}", "")
            depth = depth + open - close
        elseif l:match('"Anchor":%s*{') then
            depth = 1
        end
    end
end

function config.load(path)
    local cfg = { name = "Player", room = "rhh-ports" }
    local lines = readLines(path)
    if not lines then return cfg end
    eachLine(lines, function(l, inAnchor)
        if inAnchor then
            cfg.name = l:match('"Name":%s*"([^"]*)"') or cfg.name
            cfg.room = l:match('"RoomId":%s*"([^"]*)"') or cfg.room
        end
        return l
    end)
    return cfg
end

-- name and room must already be limited to characters that need no JSON
-- escaping (the on-screen keyboard only offers [A-Za-z0-9_-]).
function config.save(path, name, room)
    local lines = readLines(path)
    if not lines then return false, "cannot read " .. path end
    local hasNav = false
    eachLine(lines, function(l, inAnchor)
        if inAnchor then
            l = l:gsub('("Name":%s*)"[^"]*"', '%1"' .. name .. '"')
            l = l:gsub('("RoomId":%s*)"[^"]*"', '%1"' .. room .. '"')
        end
        l = l:gsub('("Menu":%s*)1(%s*,?%s*)$', '%10%2')
        if l:match('"ControlNav":') then
            hasNav = true
            l = l:gsub('("ControlNav":%s*)%d+', '%11')
        end
        return l
    end)
    if not hasNav then
        for i, l in ipairs(lines) do
            local indent = l:match('^(%s*)"gSettings":%s*{%s*$')
            if indent then
                local last = (lines[i + 1] or ""):match("^%s*}")
                table.insert(lines, i + 1, indent .. '    "ControlNav": 1' .. (last and "" or ","))
                break
            end
        end
    end
    local tmp = path .. ".tmp"
    local f, err = io.open(tmp, "w")
    if not f then return false, err end
    f:write(table.concat(lines, "\n"), "\n")
    f:close()
    return os.rename(tmp, path)
end

return config
