-- Shared copy: recipes/templates/includes/launcher/
-- Launcher config for HarbourMasters ports with Anchor online co-op (Lighthouse,
-- Ship of Harkinian). A port's launcher/config.lua is just:
--   return require("anchor")("Game Title")
--
-- Edits the game's JSON config line by line; the game writes it pretty-printed
-- (one key per line), so no JSON parser is needed:
--   CVars.gRemote.Anchor.Name / RoomId  - set (the section is created if missing)
--   CVars.gOpenWindows.Menu             - forced to 0 (menu closed at start)
--   CVars.gSettings.ControlNav          - forced to 1 (controller menu navigation)

return function(title)
    local config = {
        title = title,
        subtitle = "Anchor online co-op setup",
        help = "Players with the same Room ID play together. The name is what other players see.",
        items = {
            { id = "name", label = "Player name", kind = "text", maxlen = 20 },
            { id = "room", label = "Room ID", kind = "text", maxlen = 20 },
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

    -- Insert block (literal lines) after line i, an object's opening line, one
    -- indent level deeper; the last line gets a comma unless the object was empty.
    local function insertAfter(lines, i, block)
        local indent = lines[i]:match("^(%s*)") .. "    "
        local empty = (lines[i + 1] or ""):match("^%s*}")
        for n, e in ipairs(block) do
            local comma = (n == #block and not empty) and "," or ""
            table.insert(lines, i + n, indent .. e .. comma)
        end
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

    local function find(lines, pattern)
        for i, l in ipairs(lines) do if l:match(pattern) then return i end end
    end

    function config.load(path)
        local v = { name = "Player", room = "PORTMASTER" }
        local lines = readLines(path)
        if not lines then return v end
        eachLine(lines, function(l, inAnchor)
            if inAnchor then
                v.name = l:match('"Name":%s*"([^"]*)"') or v.name
                v.room = l:match('"RoomId":%s*"([^"]*)"') or v.room
            end
            return l
        end)
        return v
    end

    -- Names are limited by the on-screen keyboard to [A-Za-z0-9_-], so they
    -- need no JSON escaping.
    function config.save(path, v)
        local lines = readLines(path)
        if not lines then return false, "cannot read " .. path end
        local name, room = '"Name": "' .. v.name .. '"', '"RoomId": "' .. v.room .. '"'

        local hasName, hasRoom, hasNav = false, false, false
        eachLine(lines, function(l, inAnchor)
            if inAnchor then
                if l:match('"Name":') then hasName = true end
                if l:match('"RoomId":') then hasRoom = true end
                l = l:gsub('"Name":%s*"[^"]*"', name)
                l = l:gsub('"RoomId":%s*"[^"]*"', room)
            end
            l = l:gsub('("Menu":%s*)1(%s*,?%s*)$', '%10%2')
            if l:match('"ControlNav":') then
                hasNav = true
                l = l:gsub('("ControlNav":%s*)%d+', '%11')
            end
            return l
        end)

        local missing = {}
        if not hasName then missing[#missing + 1] = name end
        if not hasRoom then missing[#missing + 1] = room end
        if #missing == 2 then missing[1] = missing[1] .. "," end
        if #missing > 0 then
            local anchor = find(lines, '"Anchor":%s*{%s*$')
            local remote = find(lines, '"gRemote":%s*{%s*$')
            local cvars = find(lines, '"CVars":%s*{%s*$')
            if anchor then
                insertAfter(lines, anchor, missing)
            elseif remote then
                insertAfter(lines, remote, { '"Anchor": {', "    " .. name .. ",", "    " .. room, "}" })
            elseif cvars then
                insertAfter(lines, cvars, { '"gRemote": {', '    "Anchor": {', "        " .. name .. ",",
                                            "        " .. room, "    }", "}" })
            end
        end
        if not hasNav then
            local settings = find(lines, '"gSettings":%s*{%s*$')
            if settings then insertAfter(lines, settings, { '"ControlNav": 1' }) end
        end

        local tmp = path .. ".tmp"
        local f, err = io.open(tmp, "w")
        if not f then return false, err end
        f:write(table.concat(lines, "\n"), "\n")
        f:close()
        return os.rename(tmp, path)
    end

    return config
end
