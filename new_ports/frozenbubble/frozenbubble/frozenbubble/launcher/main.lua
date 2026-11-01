-- Frozen Bubble - PortMaster network / multiplayer setup
--
-- usage: love launcher <conf_dir> <launch_file> [device_ip]
--   conf_dir    folder holding the game's settings.ini (bound to ~/.local/share/frozen-bubble)
--   launch_file shell file written on exit, sourced by the port script
--   device_ip   shown to the player when hosting a server

local ini = require("ini")
local osk = require("osk")

local VW, VH = 640, 480
local MAX_PLAYERS = 5

local colors = {
    bg     = { 0.05, 0.10, 0.22 },
    panel  = { 0.09, 0.17, 0.33, 0.97 },
    field  = { 0.03, 0.07, 0.16 },
    key    = { 0.16, 0.27, 0.46 },
    row    = { 0.12, 0.22, 0.40 },
    accent = { 1.00, 0.82, 0.25 },
    text   = { 0.95, 0.97, 1.00 },
    dim    = { 0.62, 0.70, 0.84 },
    warn   = { 1.00, 0.45, 0.40 },
    good   = { 0.45, 0.90, 0.55 },
}
local bubbleColors = {
    { 0.30, 0.55, 1.00 }, { 1.00, 0.35, 0.35 }, { 0.40, 0.90, 0.40 },
    { 1.00, 0.85, 0.30 }, { 0.80, 0.45, 1.00 }, { 1.00, 0.60, 0.20 },
    { 0.40, 0.95, 0.95 }, { 0.95, 0.95, 0.95 },
}

local fonts = {}
local paths = {}
local deviceIP = ""
local settings, launcherIni
local cfg = {}
local screen = "main"
local sel = { main = 1, servers = 1, players = 1 }
local keyboard = nil      -- active on-screen keyboard
local toast, toastTime = nil, 0
local bubbles = {}
local transform = { scale = 1, ox = 0, oy = 0 }

---------------------------------------------------------------------------
-- Key bindings written to settings.ini
---------------------------------------------------------------------------
-- Keyboard defaults, identical to GameSettings::LoadDefaultKeys()
local KEYBOARD_KEYS = {
    { left = 80, right = 79, fire = 82, center = 81 }, -- arrows
    { left = 27, right = 25, fire = 6,  center = 7  }, -- X V C D
    { left = 4,  right = 7,  fire = 26, center = 22 }, -- A D W S
    { left = 13, right = 15, fire = 12, center = 14 }, -- J L I K
    { left = 92, right = 94, fire = 96, center = 93 }, -- keypad 4 6 8 5
}
-- Virtual controller scancodes: 300 + slot * 20 + SDL_GameControllerButton
local CTRL_SC_BASE = 300
local BTN = { a = 0, b = 1, x = 2, y = 3, up = 11, down = 12, left = 13, right = 14 }
local FIRE_CHOICES = { "a", "up", "b", "x", "y" }
local FIRE_LABEL = { a = "A", up = "D-pad Up", b = "B", x = "X", y = "Y" }

local function keysFor(p, idx)
    if p.input == "kb" then return KEYBOARD_KEYS[idx] end
    local slot = tonumber(p.input:match("pad(%d)")) or 1
    local base = CTRL_SC_BASE + (slot - 1) * 20
    local fireBtn = BTN[p.fire] or BTN.a
    return { left = base + BTN.left, right = base + BTN.right, fire = base + fireBtn, center = base + BTN.down }
end

local function inputLabel(input)
    if input == "kb" then return "Keyboard" end
    local n = input:match("pad(%d)")
    if n == "1" then return "Controller 1 (built-in)" end
    return "Controller " .. n
end

local INPUT_CHOICES = { "pad1", "pad2", "pad3", "pad4", "pad5", "kb" }

---------------------------------------------------------------------------
-- Helpers
---------------------------------------------------------------------------
local function cycle(list, value, dir)
    local idx = 1
    for i, v in ipairs(list) do if v == value then idx = i end end
    idx = (idx - 1 + dir) % #list + 1
    return list[idx]
end

local function clamp(v, lo, hi) return math.max(lo, math.min(hi, v)) end

local function showToast(msg) toast, toastTime = msg, 2.5 end

local function sanitize(s, pattern, maxlen)
    s = tostring(s or ""):gsub(pattern, "")
    return s:sub(1, maxlen)
end

local function fileExists(p)
    local f = io.open(p, "r")
    if f then f:close() return true end
    return false
end

local function currentServer()
    return cfg.servers[cfg.serverIdx]
end

local function serverLabel(s)
    if not s then return "(none - add one)" end
    return s.host .. ":" .. s.port
end

---------------------------------------------------------------------------
-- Load / save
---------------------------------------------------------------------------
local function writeGameDefaults(t)
    -- Same defaults the game writes in CreateDefaultSettings(), so a fresh
    -- file created here does not lose any settings.
    t:set("GFX", "Quality", "1")
    t:set("GFX", "LinearScaling", "false")
    t:set("GFX", "Fullscreen", "true")
    t:set("GFX", "WindowWidth", "640")
    t:set("GFX", "WindowHeight", "480")
    t:set("GFX", "ColorblindBubbles", "false")
    t:set("Sound", "EnableMusic", "true")
    t:set("Sound", "EnableSFX", "true")
    t:set("Sound", "ClassicAF", "false")
    t:set("Keys", "SpeedMultiplier", "2.0")
end

local function loadConfig()
    settings = ini.load(paths.settings)
    launcherIni = ini.load(paths.launcher)

    cfg.mode = launcherIni:get("Launcher", "Mode", "offline")
    if cfg.mode ~= "offline" and cfg.mode ~= "join" and cfg.mode ~= "host" then cfg.mode = "offline" end

    cfg.nick = sanitize(settings:get("Keys", "Nickname", ""), "[^%w_%-]", 10)
    if cfg.nick == "" then cfg.nick = "Player" end

    cfg.servers = {}
    local count = launcherIni:getnumber("Servers", "Count", 0)
    for i = 1, count do
        local v = launcherIni:get("Servers", "Server" .. i, nil)
        if v then
            local h, p = v:match("^(.-):(%d+)$")
            if h and h ~= "" then table.insert(cfg.servers, { host = h, port = tonumber(p) }) end
        end
    end
    -- Seed with whatever the game last used, if it isn't a loopback address
    local gameHost = settings:get("Network", "Host", nil)
    if #cfg.servers == 0 and gameHost and gameHost ~= "127.0.0.1" then
        table.insert(cfg.servers, { host = gameHost, port = settings:getnumber("Network", "Port", 1511) })
    end
    cfg.serverIdx = clamp(launcherIni:getnumber("Servers", "Selected", 1), 1, math.max(1, #cfg.servers))

    cfg.hostPort = clamp(launcherIni:getnumber("Host", "Port", 1511), 1024, 65535)
    cfg.hostMax = clamp(launcherIni:getnumber("Host", "MaxPlayers", 5), 2, 32)
    -- fb-server only accepts [a-zA-Z0-9.-] for -n
    cfg.hostName = sanitize(launcherIni:get("Host", "Name", "PortMaster"), "[^%w%-%.]", 24)
    if cfg.hostName == "" then cfg.hostName = "PortMaster" end
    cfg.hostLan = launcherIni:getbool("Host", "LAN", true)

    cfg.players = {}
    for i = 1, MAX_PLAYERS do
        local input = launcherIni:get("Players", "P" .. i .. "Input", "pad" .. i)
        local okInput = false
        for _, v in ipairs(INPUT_CHOICES) do if v == input then okInput = true end end
        if not okInput then input = "pad" .. i end
        local fire = launcherIni:get("Players", "P" .. i .. "Fire", "a")
        if not FIRE_LABEL[fire] then fire = "a" end
        cfg.players[i] = { input = input, fire = fire }
    end
end

local function saveConfig()
    -- launcher state
    launcherIni:set("Launcher", "Mode", cfg.mode)
    launcherIni:set("Servers", "Count", #cfg.servers)
    for i, s in ipairs(cfg.servers) do
        launcherIni:set("Servers", "Server" .. i, s.host .. ":" .. s.port)
    end
    -- drop stale entries
    local sec = launcherIni.sections["servers"]
    if sec then
        local keep = {}
        for _, lk in ipairs(sec.keys) do
            local n = tonumber(lk:match("^server(%d+)$"))
            if n and n > #cfg.servers then sec.values[lk] = nil else table.insert(keep, lk) end
        end
        sec.keys = keep
    end
    launcherIni:set("Servers", "Selected", cfg.serverIdx)
    launcherIni:set("Host", "Port", cfg.hostPort)
    launcherIni:set("Host", "MaxPlayers", cfg.hostMax)
    launcherIni:set("Host", "Name", cfg.hostName)
    launcherIni:set("Host", "LAN", cfg.hostLan)
    for i, p in ipairs(cfg.players) do
        launcherIni:set("Players", "P" .. i .. "Input", p.input)
        launcherIni:set("Players", "P" .. i .. "Fire", p.fire)
    end
    local ok1, err1 = launcherIni:save()

    -- game settings
    if not fileExists(paths.settings) then writeGameDefaults(settings) end
    settings:set("GFX", "Fullscreen", "true")
    settings:set("Keys", "Nickname", cfg.nick)
    for i, p in ipairs(cfg.players) do
        local k = keysFor(p, i)
        settings:set("Keys", "P" .. i .. "Left", k.left)
        settings:set("Keys", "P" .. i .. "Right", k.right)
        settings:set("Keys", "P" .. i .. "Fire", k.fire)
        settings:set("Keys", "P" .. i .. "Center", k.center)
    end
    if cfg.mode == "join" and currentServer() then
        settings:set("Network", "Host", currentServer().host)
        settings:set("Network", "Port", currentServer().port)
    elseif cfg.mode == "host" then
        settings:set("Network", "Host", "127.0.0.1")
        settings:set("Network", "Port", cfg.hostPort)
    else
        settings:set("Network", "Host", "")
        settings:set("Network", "Port", 1511)
    end
    local ok2, err2 = settings:save()

    if not ok1 or not ok2 then
        print("launcher: failed to save settings: " .. tostring(err1 or err2))
    end
end

local function writeLaunchFile(launch)
    local hosting = launch and cfg.mode == "host"
    local lines = {
        "# Written by the Frozen Bubble launcher",
        "FB_LAUNCH=" .. (launch and 1 or 0),
        "FB_MODE=" .. cfg.mode,
        "FB_HOST_SERVER=" .. (hosting and 1 or 0),
        "FB_SERVER_PORT=" .. cfg.hostPort,
        "FB_SERVER_MAX=" .. cfg.hostMax,
        "FB_SERVER_NAME='" .. cfg.hostName:gsub("'", "") .. "'",
        "FB_SERVER_LAN=" .. (cfg.hostLan and 1 or 0),
        "",
    }
    local f = io.open(paths.launch, "w")
    if f then
        f:write(table.concat(lines, "\n"))
        f:close()
    else
        print("launcher: cannot write " .. tostring(paths.launch))
    end
end

local function finish(launch)
    saveConfig()
    writeLaunchFile(launch)
    love.event.quit(0)
end

---------------------------------------------------------------------------
-- Menus
---------------------------------------------------------------------------
local MODES = { "offline", "join", "host" }
local MODE_LABEL = { offline = "Offline / local", join = "Join a server", host = "Host a server" }

local function openKeyboard(opts, cb)
    toast = nil
    keyboard = osk.new(opts, function(ok, value)
        keyboard = nil
        if ok then cb(value) end
    end)
end

local function editServer(index)
    local s = cfg.servers[index] or { host = "", port = 1511 }
    openKeyboard({ title = "Server address (IP or hostname)", value = s.host, kind = "host", maxlen = 63,
                   allowed = "[%w%.%-:_]" }, function(host)
        host = sanitize(host, "[^%w%.%-_]", 63)
        if host == "" then showToast("Address cannot be empty") return end
        openKeyboard({ title = "Server port", value = tostring(s.port), kind = "number", maxlen = 5,
                       allowed = "%d" }, function(port)
            port = tonumber(port) or 1511
            if port < 1 or port > 65535 then port = 1511 end
            local entry = { host = host, port = port }
            if cfg.servers[index] then
                cfg.servers[index] = entry
            else
                table.insert(cfg.servers, entry)
                index = #cfg.servers
                screen = "main"
            end
            cfg.serverIdx = index
            showToast("Saved " .. serverLabel(entry))
        end)
    end)
end

local function mainItems()
    local items = {
        { id = "play", label = "Play Frozen Bubble" },
        { id = "mode", label = "Network mode", value = MODE_LABEL[cfg.mode], adjust = true },
        { id = "nick", label = "Nickname", value = cfg.nick },
    }
    if cfg.mode == "join" then
        table.insert(items, { id = "server", label = "Server", value = serverLabel(currentServer()),
                              adjust = #cfg.servers > 1 })
    elseif cfg.mode == "host" then
        table.insert(items, { id = "hport", label = "Server port", value = tostring(cfg.hostPort), adjust = true })
        table.insert(items, { id = "hmax", label = "Max connected players", value = tostring(cfg.hostMax), adjust = true })
        table.insert(items, { id = "hname", label = "Server name", value = cfg.hostName })
        table.insert(items, { id = "hlan", label = "LAN discovery", value = cfg.hostLan and "On" or "Off", adjust = true })
    end
    local summary = {}
    for i = 1, 2 do
        local p = cfg.players[i]
        table.insert(summary, "P" .. i .. " " .. (p.input == "kb" and "Kbd" or ("Pad " .. p.input:sub(4))))
    end
    table.insert(items, { id = "players", label = "Local players / controllers", value = table.concat(summary, ", ") .. " ..." })
    table.insert(items, { id = "quit", label = "Save and exit" })
    return items
end

local function modeHelp()
    if cfg.mode == "offline" then
        return "Single player and local 2-player. In-game Net/LAN Game menus still work."
    elseif cfg.mode == "join" then
        if not currentServer() then return "Add a server first: select Server and press A." end
        return "In game: Net Game > Manual entry is pre-filled with this server - press A twice to connect."
    else
        local ip = deviceIP ~= "" and deviceIP or "this device's IP"
        return "A server starts before the game. Friends join " .. ip .. ":" .. cfg.hostPort ..
               ". In game use LAN Game, or Net Game > Manual entry > Connect."
    end
end

local function mainActivate(item, action)
    local id = item.id
    local dir = (action == "left") and -1 or ((action == "right") and 1 or 0)
    local big = (action == "lb") and -100 or ((action == "rb") and 100 or 0)

    if id == "play" and action == "a" then
        if cfg.mode == "join" and not currentServer() then
            showToast("Add a server first")
            return
        end
        finish(true)
    elseif id == "quit" and action == "a" then
        finish(false)
    elseif id == "mode" and (dir ~= 0 or action == "a") then
        cfg.mode = cycle(MODES, cfg.mode, dir ~= 0 and dir or 1)
    elseif id == "nick" and action == "a" then
        openKeyboard({ title = "Nickname (max 10: letters, digits, - _)", value = cfg.nick, kind = "nick",
                       maxlen = 10, allowed = "[%w_%-]" }, function(v)
            v = sanitize(v, "[^%w_%-]", 10)
            if v == "" then showToast("Nickname cannot be empty") return end
            cfg.nick = v
        end)
    elseif id == "server" then
        if action == "a" then
            screen = "servers"
            sel.servers = clamp(cfg.serverIdx, 1, #cfg.servers + 1)
        elseif dir ~= 0 and #cfg.servers > 0 then
            cfg.serverIdx = (cfg.serverIdx - 1 + dir) % #cfg.servers + 1
        end
    elseif id == "hport" then
        if dir ~= 0 or big ~= 0 then
            cfg.hostPort = clamp(cfg.hostPort + dir + big, 1024, 65535)
        elseif action == "a" then
            openKeyboard({ title = "Server port", value = tostring(cfg.hostPort), kind = "number", maxlen = 5,
                           allowed = "%d" }, function(v)
                cfg.hostPort = clamp(tonumber(v) or 1511, 1024, 65535)
            end)
        end
    elseif id == "hmax" and dir ~= 0 then
        cfg.hostMax = clamp(cfg.hostMax + dir, 2, 32)
    elseif id == "hname" and action == "a" then
        openKeyboard({ title = "Server name (letters, digits, . -)", value = cfg.hostName, kind = "host",
                       maxlen = 24, allowed = "[%w%-%.]" }, function(v)
            v = sanitize(v, "[^%w%-%.]", 24)
            if v ~= "" then cfg.hostName = v end
        end)
    elseif id == "hlan" and (dir ~= 0 or action == "a") then
        cfg.hostLan = not cfg.hostLan
    elseif id == "players" and action == "a" then
        screen = "players"
        sel.players = 1
    end
end

local function playerItems()
    local items = {}
    for i = 1, MAX_PLAYERS do table.insert(items, { id = "p", idx = i }) end
    table.insert(items, { id = "reset", label = "Reset to defaults" })
    table.insert(items, { id = "back", label = "Back" })
    return items
end

local function playerConflicts()
    local used, dup = {}, {}
    for i, p in ipairs(cfg.players) do
        if used[p.input] and p.input ~= "kb" then
            dup[i] = true; dup[used[p.input]] = true
        end
        used[p.input] = used[p.input] or i
    end
    return dup
end

local function serverItems()
    local items = {}
    for i, s in ipairs(cfg.servers) do table.insert(items, { id = "srv", idx = i, label = serverLabel(s) }) end
    table.insert(items, { id = "add", label = "+ Add server" })
    table.insert(items, { id = "back", label = "Back" })
    return items
end

---------------------------------------------------------------------------
-- Input
---------------------------------------------------------------------------
local function handleAction(action)
    if keyboard then keyboard:action(action) return end

    if screen == "main" then
        local items = mainItems()
        sel.main = clamp(sel.main, 1, #items)
        if action == "up" then sel.main = (sel.main - 2) % #items + 1
        elseif action == "down" then sel.main = sel.main % #items + 1
        elseif action == "start" then mainActivate({ id = "play" }, "a")
        elseif action == "b" or action == "back" then
            -- nothing: B on the main menu does not quit, to avoid accidental exits
        else
            mainActivate(items[sel.main], action)
        end

    elseif screen == "servers" then
        local items = serverItems()
        sel.servers = clamp(sel.servers, 1, #items)
        local it = items[sel.servers]
        if action == "up" then sel.servers = (sel.servers - 2) % #items + 1
        elseif action == "down" then sel.servers = sel.servers % #items + 1
        elseif action == "b" or action == "back" then screen = "main"
        elseif action == "a" then
            if it.id == "srv" then cfg.serverIdx = it.idx; screen = "main"
            elseif it.id == "add" then editServer(#cfg.servers + 1)
            else screen = "main" end
        elseif action == "x" and it.id == "srv" then
            editServer(it.idx)
        elseif action == "y" and it.id == "srv" then
            table.remove(cfg.servers, it.idx)
            if cfg.serverIdx > #cfg.servers then cfg.serverIdx = math.max(1, #cfg.servers) end
            showToast("Server removed")
        end

    elseif screen == "players" then
        local items = playerItems()
        sel.players = clamp(sel.players, 1, #items)
        local it = items[sel.players]
        local dir = (action == "left") and -1 or ((action == "right") and 1 or 0)
        if action == "up" then sel.players = (sel.players - 2) % #items + 1
        elseif action == "down" then sel.players = sel.players % #items + 1
        elseif action == "b" or action == "back" then screen = "main"
        elseif it.id == "p" then
            local p = cfg.players[it.idx]
            if dir ~= 0 or action == "a" then
                p.input = cycle(INPUT_CHOICES, p.input, dir ~= 0 and dir or 1)
            elseif action == "x" then
                p.fire = cycle(FIRE_CHOICES, p.fire, 1)
            end
        elseif it.id == "reset" and action == "a" then
            for i = 1, MAX_PLAYERS do cfg.players[i] = { input = "pad" .. i, fire = "a" } end
            showToast("Controllers reset")
        elseif it.id == "back" and action == "a" then
            screen = "main"
        end
    end
end

-- gamepad repeat for held directions
local held = { action = nil, t = 0 }
local REPEAT_DELAY, REPEAT_RATE = 0.35, 0.08

local function press(action, repeatable)
    handleAction(action)
    if repeatable then held.action, held.t = action, REPEAT_DELAY end
end

local function release(action)
    if held.action == action then held.action = nil end
end

local PAD_MAP = {
    dpup = "up", dpdown = "down", dpleft = "left", dpright = "right",
    a = "a", b = "b", x = "x", y = "y", start = "start", back = "back",
    leftshoulder = "lb", rightshoulder = "rb",
}
local DIRS = { up = true, down = true, left = true, right = true, lb = true, rb = true }

function love.gamepadpressed(_, button)
    local a = PAD_MAP[button]
    if a then press(a, DIRS[a]) end
end

function love.gamepadreleased(_, button)
    local a = PAD_MAP[button]
    if a then release(a) end
end

local axisState = {}
function love.gamepadaxis(js, axis, value)
    local id = js:getID()
    axisState[id] = axisState[id] or {}
    local st = axisState[id]
    local dead = 0.5
    if axis == "leftx" or axis == "lefty" then
        local neg = (axis == "leftx") and "left" or "up"
        local pos = (axis == "leftx") and "right" or "down"
        local want = (value < -dead) and neg or ((value > dead) and pos or nil)
        if st[axis] ~= want then
            if st[axis] then release(st[axis]) end
            if want then press(want, true) end
            st[axis] = want
        end
    end
end

local KEY_MAP = {
    up = "up", down = "down", left = "left", right = "right",
    ["return"] = "a", kpenter = "a", space = "a",
    escape = "b", backspace = "b",
    pageup = "lb", pagedown = "rb",
    tab = "x", delete = "y",
    f1 = "start",
}
local OSK_KEY_MAP = {
    up = "up", down = "down", left = "left", right = "right",
    ["return"] = "start", kpenter = "start", backspace = "b", escape = "back",
}

function love.keypressed(key)
    if keyboard then
        local a = OSK_KEY_MAP[key]
        if a then handleAction(a) end
        return
    end
    local a = KEY_MAP[key]
    if a then handleAction(a) end
end

function love.textinput(t)
    if keyboard then keyboard:textinput(t) end
end

---------------------------------------------------------------------------
-- Love callbacks
---------------------------------------------------------------------------
local function updateTransform()
    local w, h = love.graphics.getDimensions()
    local s = math.min(w / VW, h / VH)
    transform.scale = s
    transform.ox = math.floor((w - VW * s) / 2)
    transform.oy = math.floor((h - VH * s) / 2)
end

function love.load(args)
    paths.conf = args[1] or "."
    paths.launch = args[2] or (paths.conf .. "/launch.cfg")
    deviceIP = args[3] or ""
    paths.settings = paths.conf .. "/settings.ini"
    paths.launcher = paths.conf .. "/launcher.ini"

    love.keyboard.setKeyRepeat(true)
    love.graphics.setDefaultFilter("linear", "linear")
    updateTransform()
    -- fonts are created at the final pixel size so they stay sharp when scaled
    local s = transform.scale
    local function f(px)
        local font = love.graphics.newFont(math.max(8, math.floor(px * s + 0.5)))
        return font
    end
    fonts.small, fonts.normal, fonts.big, fonts.title, fonts.logo = f(13), f(17), f(22), f(22), f(40)
    fonts.scale = s

    loadConfig()

    for i = 1, 18 do
        bubbles[i] = {
            x = love.math.random(0, VW), y = love.math.random(0, VH),
            r = love.math.random(6, 18), v = love.math.random(8, 25),
            c = bubbleColors[love.math.random(1, #bubbleColors)],
        }
    end
    -- default selection is "Play"
    sel.main = 1
end

function love.resize()
    updateTransform()
end

function love.update(dt)
    if held.action then
        held.t = held.t - dt
        if held.t <= 0 then
            handleAction(held.action)
            held.t = REPEAT_RATE
        end
    end
    if toast then
        toastTime = toastTime - dt
        if toastTime <= 0 then toast = nil end
    end
    for _, b in ipairs(bubbles) do
        b.y = b.y - b.v * dt
        if b.y < -20 then b.y = VH + 20; b.x = love.math.random(0, VW) end
    end
end

-- Text is drawn unscaled at screen resolution for sharpness; everything
-- else is laid out in a 640x480 virtual space.
local function vprint(text, font, x, y, w, align)
    local s = transform.scale
    love.graphics.push()
    love.graphics.origin()
    love.graphics.translate(transform.ox, transform.oy)
    love.graphics.setFont(font)
    love.graphics.printf(text, math.floor(x * s), math.floor(y * s), math.floor(w * s), align or "left")
    love.graphics.pop()
end

local function drawRow(y, label, value, selected, adjust, valueColor)
    if selected then
        love.graphics.setColor(colors.accent[1], colors.accent[2], colors.accent[3], 0.18)
        love.graphics.rectangle("fill", 40, y - 3, VW - 80, 28, 8, 8)
        love.graphics.setColor(colors.accent)
        love.graphics.rectangle("line", 40, y - 3, VW - 80, 28, 8, 8)
    end
    love.graphics.setColor(selected and colors.accent or colors.text)
    vprint(label, fonts.normal, 56, y + 2, 300, "left")
    if value then
        local v = adjust and ("<  " .. value .. "  >") or value
        love.graphics.setColor(valueColor or (selected and colors.text or colors.dim))
        vprint(v, fonts.normal, 300, y + 2, VW - 356, "right")
    end
end

local function drawHeader(subtitle)
    -- logo
    for i = 1, 8 do
        local c = bubbleColors[i]
        love.graphics.setColor(c[1], c[2], c[3], 0.9)
        love.graphics.circle("fill", 150 + i * 38, 26, 10)
        love.graphics.setColor(1, 1, 1, 0.5)
        love.graphics.circle("fill", 147 + i * 38, 22, 3)
    end
    love.graphics.setColor(colors.text)
    vprint("Frozen Bubble", fonts.logo, 0, 40, VW, "center")
    love.graphics.setColor(colors.dim)
    vprint(subtitle, fonts.small, 0, 90, VW, "center")
end

local function drawFooter(text)
    love.graphics.setColor(0, 0, 0, 0.35)
    love.graphics.rectangle("fill", 0, VH - 30, VW, 30)
    love.graphics.setColor(colors.dim)
    vprint(text, fonts.small, 0, VH - 23, VW, "center")
end

local function drawMain()
    drawHeader("Network & multiplayer setup")
    local items = mainItems()
    sel.main = clamp(sel.main, 1, #items)
    local y = 118
    for i, it in ipairs(items) do
        local vc = nil
        if it.id == "server" and not currentServer() then vc = colors.warn end
        if it.id == "play" then
            local s = (i == sel.main)
            love.graphics.setColor(s and colors.accent or colors.row)
            love.graphics.rectangle("fill", 160, y - 4, VW - 320, 34, 17, 17)
            love.graphics.setColor(s and colors.bg or colors.text)
            vprint(it.label, fonts.big, 160, y + 1, VW - 320, "center")
            y = y + 44
        else
            drawRow(y, it.label, it.value, i == sel.main, it.adjust, vc)
            y = y + 30
        end
    end

    love.graphics.setColor(colors.dim)
    vprint(modeHelp(), fonts.small, 40, VH - 70, VW - 80, "center")

    local it = items[sel.main]
    local hint = "A select   Up/Down move   Start play"
    if it.adjust then hint = "Left/Right change   " .. hint end
    if it.id == "hport" then hint = "Left/Right +/-1   L1/R1 +/-100   A type   Start play" end
    if it.id == "server" then hint = "A manage servers   Left/Right switch   Start play" end
    drawFooter(hint)
end

local function drawServers()
    drawHeader("Saved servers")
    local items = serverItems()
    sel.servers = clamp(sel.servers, 1, #items)
    local y = 118
    local first = math.max(1, sel.servers - 7)
    for i = first, math.min(#items, first + 8) do
        local it = items[i]
        local value = nil
        if it.id == "srv" and it.idx == cfg.serverIdx then value = "in use" end
        drawRow(y, it.label, value, i == sel.servers, false, colors.good)
        y = y + 34
    end
    local it = items[sel.servers]
    if it.id == "srv" then
        drawFooter("A use   X edit   Y delete   B back")
    else
        drawFooter("A select   B back")
    end
end

local function drawPlayers()
    drawHeader("Local players / controllers")
    local items = playerItems()
    sel.players = clamp(sel.players, 1, #items)
    local dup = playerConflicts()
    local y = 112
    for i, it in ipairs(items) do
        local s = (i == sel.players)
        if it.id == "p" then
            local p = cfg.players[it.idx]
            if s then
                love.graphics.setColor(colors.accent[1], colors.accent[2], colors.accent[3], 0.18)
                love.graphics.rectangle("fill", 40, y - 4, VW - 80, 32, 8, 8)
                love.graphics.setColor(colors.accent)
                love.graphics.rectangle("line", 40, y - 4, VW - 80, 32, 8, 8)
            end
            local c = bubbleColors[it.idx]
            love.graphics.setColor(c)
            love.graphics.circle("fill", 62, y + 12, 9)
            love.graphics.setColor(s and colors.accent or colors.text)
            vprint("Player " .. it.idx, fonts.normal, 80, y + 2, 120, "left")
            love.graphics.setColor(dup[it.idx] and colors.warn or (s and colors.text or colors.dim))
            local v = inputLabel(p.input)
            vprint(s and ("<  " .. v .. "  >") or v, fonts.normal, 190, y + 2, 270, "center")
            love.graphics.setColor(s and colors.text or colors.dim)
            local fireText = (p.input == "kb") and "Fire: default" or ("Fire: " .. FIRE_LABEL[p.fire])
            vprint(fireText, fonts.small, 460, y + 5, 130, "right")
            y = y + 34
        else
            drawRow(y, it.label, nil, s)
            y = y + 34
        end
    end

    love.graphics.setColor(colors.dim)
    local note = "Controller 1 is the built-in pad; extra USB/Bluetooth pads are numbered in the order " ..
                 "they connect. Upstream supports 2 local players; 3-5 is experimental."
    if next(dup) then
        love.graphics.setColor(colors.warn)
        note = "Two players share the same controller - they will move together."
    end
    vprint(note, fonts.small, 40, VH - 70, VW - 80, "center")
    local it = items[sel.players]
    if it.id == "p" then
        drawFooter("Left/Right controller   X fire button   B back")
    else
        drawFooter("A select   B back")
    end
end

function love.draw()
    love.graphics.clear(colors.bg)
    love.graphics.push()
    love.graphics.translate(transform.ox, transform.oy)
    love.graphics.scale(transform.scale)
    -- clip to virtual area
    love.graphics.setScissor(transform.ox, transform.oy, VW * transform.scale, VH * transform.scale)

    for _, b in ipairs(bubbles) do
        love.graphics.setColor(b.c[1], b.c[2], b.c[3], 0.12)
        love.graphics.circle("fill", b.x, b.y, b.r)
    end

    if screen == "main" then drawMain()
    elseif screen == "servers" then drawServers()
    elseif screen == "players" then drawPlayers() end

    if keyboard then
        love.graphics.setColor(0, 0, 0, 0.6)
        love.graphics.rectangle("fill", 0, 0, VW, VH)
        keyboard:draw(fonts, colors, vprint)
    end

    if toast then
        love.graphics.setColor(0, 0, 0, 0.85)
        love.graphics.rectangle("fill", 120, 84, VW - 240, 28, 8, 8)
        love.graphics.setColor(colors.accent)
        vprint(toast, fonts.normal, 120, 88, VW - 240, "center")
    end

    love.graphics.setScissor()
    love.graphics.pop()
end
