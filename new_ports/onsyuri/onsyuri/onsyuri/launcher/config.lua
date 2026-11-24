-- ONScripter-Yuri launcher: picks a game folder in games/ and the script encoding,
-- saved as shell variables in onsyuri.cfg (sourced by ONScripter-Yuri.sh)

local GAMES = (love.filesystem.getSource():match("^(.*)/[^/]+/?$") or ".") .. "/games"

local function quote(s) return "'" .. s:gsub("'", "'\\''") .. "'" end

-- Every folder in games/ is a game
local games = {}
local p = io.popen("ls -1 -p " .. quote(GAMES) .. " 2>/dev/null")
if p then
    for l in p:lines() do
        local name = l:match("^(.+)/$")
        if name then games[#games + 1] = { name, name } end
    end
    p:close()
end

local config = {
    title = "ONScripter-Yuri",
    subtitle = "NScripter visual novel engine",
    help = "A advances, B opens the menu, X skips while held.",
    items = {
        { id = "GAME", label = "Game", kind = "choice", choices = games,
          help = "Each folder in ports/onsyuri/games/ is a game." },
        { id = "ENCODING", label = "Script encoding", kind = "choice",
          choices = { { "sjis", "Shift-JIS (Japanese, English)" }, { "gbk", "GBK (Chinese)" }, { "utf8", "UTF-8" } },
          help = "Try another one if the text is garbled." },
        { id = "SCREEN", label = "Screen", kind = "choice",
          choices = { { "fit", "Keep aspect ratio" }, { "stretch", "Stretch" } } },
    },
}

if #games == 0 then
    games[1] = { "", "None found" }
    config.notice = "Copy a game into ports/onsyuri/games/<name>/"
    LAUNCHER_AUTOSTART = 0
end

local DEFAULTS = { GAME = games[1][1], ENCODING = "sjis", SCREEN = "fit" }

function config.load(path)
    local v = {}
    for k, d in pairs(DEFAULTS) do v[k] = d end
    local f = io.open(path, "r")
    if f then
        for l in f:lines() do
            local k, s = l:match("^([%u]+)='(.*)'$")
            if k and v[k] then v[k] = s:gsub("'\\''", "'") end
        end
        f:close()
    end
    -- A game that was removed falls back to the first one
    local found = false
    for _, g in ipairs(games) do found = found or g[1] == v.GAME end
    if not found then v.GAME = DEFAULTS.GAME end
    return v
end

function config.save(path, v)
    local f, err = io.open(path, "w")
    if not f then return false, err end
    for _, it in ipairs(config.items) do f:write(it.id, "=", quote(v[it.id]), "\n") end
    f:close()
    return true
end

return config
