-- Minimal INI reader/writer that keeps the layout of an existing file.
-- Section and key lookups are case-insensitive (the game uses iniparser,
-- which lowercases everything), but new keys are written as given.

local ini = {}
ini.__index = ini

local function trim(s)
    return (s:gsub("^%s+", ""):gsub("%s+$", ""))
end

function ini.load(path)
    local self = setmetatable({ path = path, sections = {}, order = {} }, ini)
    local f = io.open(path, "r")
    if not f then return self end
    local cur = self:_section("")
    for line in f:lines() do
        line = line:gsub("\r$", "")
        local name = line:match("^%s*%[([^%]]+)%]%s*$")
        if name then
            cur = self:_section(trim(name))
        else
            local k, v = line:match("^%s*([^=;#]-)%s*=%s*(.-)%s*$")
            if k and k ~= "" then
                v = v:gsub("^\"(.*)\"$", "%1")
                cur.values[k:lower()] = { key = k, value = v }
                table.insert(cur.keys, k:lower())
            end
        end
    end
    f:close()
    return self
end

function ini:_section(name)
    local lname = name:lower()
    local s = self.sections[lname]
    if not s then
        s = { name = name, keys = {}, values = {} }
        self.sections[lname] = s
        table.insert(self.order, lname)
    end
    return s
end

function ini:get(section, key, default)
    local s = self.sections[section:lower()]
    if not s then return default end
    local e = s.values[key:lower()]
    if not e or e.value == "" then return default end
    return e.value
end

function ini:getnumber(section, key, default)
    return tonumber(self:get(section, key, nil)) or default
end

function ini:getbool(section, key, default)
    local v = self:get(section, key, nil)
    if v == nil then return default end
    v = v:lower()
    return v == "true" or v == "1" or v == "yes" or v == "on"
end

function ini:set(section, key, value)
    local s = self:_section(section)
    local lk = key:lower()
    local e = s.values[lk]
    if value == true then value = "true" elseif value == false then value = "false" end
    value = tostring(value)
    if e then
        e.value = value
    else
        s.values[lk] = { key = key, value = value }
        table.insert(s.keys, lk)
    end
end

function ini:save(path)
    path = path or self.path
    local out = {}
    for _, lname in ipairs(self.order) do
        local s = self.sections[lname]
        if #s.keys > 0 or lname ~= "" then
            if lname ~= "" then table.insert(out, "[" .. s.name .. "]") end
            for _, lk in ipairs(s.keys) do
                local e = s.values[lk]
                table.insert(out, e.key .. " = " .. e.value)
            end
            table.insert(out, "")
        end
    end
    local f, err = io.open(path, "w")
    if not f then return false, err end
    f:write(table.concat(out, "\n"))
    f:close()
    return true
end

return ini
