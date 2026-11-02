-- Lighthouse - PortMaster launcher: Anchor (online co-op) player name and room
--
-- usage: love launcher <config_json> <launch_file>
--   config_json  the game's lighthouse.cfg.json (edited on Play / Save)
--   launch_file  shell file written on exit, sourced by the port script

local config = require("config")
local osk = require("osk")

local VW, VH = 640, 480

local colors = {
    bg     = { 0.06, 0.08, 0.14 },
    panel  = { 0.10, 0.14, 0.24, 0.97 },
    field  = { 0.03, 0.05, 0.10 },
    key    = { 0.17, 0.24, 0.38 },
    row    = { 0.13, 0.19, 0.32 },
    accent = { 1.00, 0.78, 0.30 },
    text   = { 0.95, 0.97, 1.00 },
    dim    = { 0.62, 0.68, 0.80 },
    warn   = { 1.00, 0.45, 0.40 },
}

local fonts = {}
local paths = {}
local cfg
local sel = 1
local keyboard = nil
local toast, toastTime = nil, 0
local transform = { scale = 1, ox = 0, oy = 0 }

local ITEMS = {
    { id = "play", label = "Play Lighthouse" },
    { id = "name", label = "Player name" },
    { id = "room", label = "Room ID" },
    { id = "quit", label = "Save and exit" },
}

local function finish(launch)
    local ok, err = config.save(paths.config, cfg.name, cfg.room)
    if not ok then print("launcher: " .. tostring(err)) end
    local f = io.open(paths.launch, "w")
    if f then
        f:write("# Written by the Lighthouse launcher\nLH_LAUNCH=" .. (launch and 1 or 0) .. "\n")
        f:close()
    end
    love.event.quit(0)
end

local function edit(field, title)
    keyboard = osk.new({ title = title, value = cfg[field], kind = "nick", maxlen = 20,
                         allowed = "[%w_%-]" }, function(ok, value)
        keyboard = nil
        if not ok then return end
        if value == "" then
            toast, toastTime = title .. " cannot be empty", 2.5
        else
            cfg[field] = value
        end
    end)
end

local function handleAction(action)
    if keyboard then keyboard:action(action) return end
    local it = ITEMS[sel]
    if action == "up" then sel = (sel - 2) % #ITEMS + 1
    elseif action == "down" then sel = sel % #ITEMS + 1
    elseif action == "start" then finish(true)
    elseif action == "back" then finish(false)
    elseif action == "a" then
        if it.id == "play" then finish(true)
        elseif it.id == "quit" then finish(false)
        elseif it.id == "name" then edit("name", "Player name")
        elseif it.id == "room" then edit("room", "Room ID") end
    end
end

---------------------------------------------------------------------------
-- Input (gamepad with held-direction repeat, plus keyboard)
---------------------------------------------------------------------------
local held = { action = nil, t = 0 }
local REPEAT_DELAY, REPEAT_RATE = 0.35, 0.08
local DIRS = { up = true, down = true, left = true, right = true }

local function press(action)
    handleAction(action)
    if DIRS[action] then held.action, held.t = action, REPEAT_DELAY end
end

local function release(action)
    if held.action == action then held.action = nil end
end

local PAD_MAP = {
    dpup = "up", dpdown = "down", dpleft = "left", dpright = "right",
    a = "a", b = "b", x = "x", y = "y", start = "start", back = "back",
}

function love.gamepadpressed(_, button)
    if PAD_MAP[button] then press(PAD_MAP[button]) end
end

function love.gamepadreleased(_, button)
    if PAD_MAP[button] then release(PAD_MAP[button]) end
end

local axisState = {}
function love.gamepadaxis(js, axis, value)
    if axis ~= "leftx" and axis ~= "lefty" then return end
    local st = axisState[js:getID()] or {}
    axisState[js:getID()] = st
    local neg = (axis == "leftx") and "left" or "up"
    local pos = (axis == "leftx") and "right" or "down"
    local want = (value < -0.5) and neg or ((value > 0.5) and pos or nil)
    if st[axis] ~= want then
        if st[axis] then release(st[axis]) end
        if want then press(want) end
        st[axis] = want
    end
end

local KEY_MAP = {
    up = "up", down = "down", left = "left", right = "right",
    ["return"] = "a", kpenter = "a", space = "a", escape = "back",
}
local OSK_KEY_MAP = {
    up = "up", down = "down", left = "left", right = "right",
    ["return"] = "start", kpenter = "start", backspace = "b", escape = "back",
}

function love.keypressed(key)
    local a = (keyboard and OSK_KEY_MAP or KEY_MAP)[key]
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
    paths.config = args[1] or "lighthouse.cfg.json"
    paths.launch = args[2] or "launch.cfg"
    love.keyboard.setKeyRepeat(true)
    updateTransform()
    -- fonts are created at the final pixel size so they stay sharp when scaled
    local function f(px) return love.graphics.newFont(math.max(8, math.floor(px * transform.scale + 0.5))) end
    fonts.small, fonts.normal, fonts.big, fonts.title, fonts.logo = f(13), f(17), f(22), f(22), f(40)
    cfg = config.load(paths.config)
end

function love.resize()
    updateTransform()
end

function love.update(dt)
    if held.action then
        held.t = held.t - dt
        if held.t <= 0 then handleAction(held.action); held.t = REPEAT_RATE end
    end
    if toast then
        toastTime = toastTime - dt
        if toastTime <= 0 then toast = nil end
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

local function drawMain()
    love.graphics.setColor(colors.text)
    vprint("Lighthouse", fonts.logo, 0, 40, VW, "center")
    love.graphics.setColor(colors.dim)
    vprint("Anchor online co-op setup", fonts.small, 0, 90, VW, "center")

    local y = 140
    for i, it in ipairs(ITEMS) do
        local s = (i == sel)
        if it.id == "play" then
            love.graphics.setColor(s and colors.accent or colors.row)
            love.graphics.rectangle("fill", 160, y - 4, VW - 320, 34, 17, 17)
            love.graphics.setColor(s and colors.bg or colors.text)
            vprint(it.label, fonts.big, 160, y + 1, VW - 320, "center")
            y = y + 50
        else
            if s then
                love.graphics.setColor(colors.accent[1], colors.accent[2], colors.accent[3], 0.18)
                love.graphics.rectangle("fill", 40, y - 3, VW - 80, 28, 8, 8)
                love.graphics.setColor(colors.accent)
                love.graphics.rectangle("line", 40, y - 3, VW - 80, 28, 8, 8)
            end
            love.graphics.setColor(s and colors.accent or colors.text)
            vprint(it.label, fonts.normal, 56, y + 2, 300, "left")
            if cfg[it.id] then
                love.graphics.setColor(s and colors.text or colors.dim)
                vprint(cfg[it.id], fonts.normal, 300, y + 2, VW - 356, "right")
            end
            y = y + 34
        end
    end

    love.graphics.setColor(colors.dim)
    vprint("Players with the same Room ID play together. The name is what other players see.",
           fonts.small, 40, VH - 80, VW - 80, "center")

    love.graphics.setColor(0, 0, 0, 0.35)
    love.graphics.rectangle("fill", 0, VH - 30, VW, 30)
    love.graphics.setColor(colors.dim)
    vprint("A select   Up/Down move   Start play   Select exit", fonts.small, 0, VH - 23, VW, "center")
end

function love.draw()
    love.graphics.clear(colors.bg)
    love.graphics.push()
    love.graphics.translate(transform.ox, transform.oy)
    love.graphics.scale(transform.scale)
    love.graphics.setScissor(transform.ox, transform.oy, VW * transform.scale, VH * transform.scale)

    drawMain()

    if keyboard then
        love.graphics.setColor(0, 0, 0, 0.6)
        love.graphics.rectangle("fill", 0, 0, VW, VH)
        keyboard:draw(fonts, colors, vprint)
    end

    if toast then
        love.graphics.setColor(0, 0, 0, 0.85)
        love.graphics.rectangle("fill", 120, 104, VW - 240, 28, 8, 8)
        love.graphics.setColor(colors.warn)
        vprint(toast, fonts.normal, 120, 108, VW - 240, "center")
    end

    love.graphics.setScissor()
    love.graphics.pop()
end
