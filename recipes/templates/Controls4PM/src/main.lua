local DEFAULT_IMAGE_PATH = "controls.png"
local DEFAULT_DURATION = 5

local converter = require("converter")

local imagePath = DEFAULT_IMAGE_PATH
local displayDuration = DEFAULT_DURATION
local targetW = nil
local targetH = nil
local scaleX = 1.0
local scaleY = 1.0
local controlsImage = nil
local elapsed = 0
local windowWidth, windowHeight = 0, 0
local imageX, imageY = 0, 0
local fontHeight = 0
local countdownFont = nil

local function showHelp()
    print("Controls Display - LÖVE 2D Fullscreen Image Viewer")
    print()
    print("Usage: love . [image_path] [duration_seconds] [scale_x] [scale_y]")
    print("       love . create -c CONFIG_FILE -i INPUT_SVG -r README_FILE [-p PORT_JSON]")
    print()
    print("Arguments:")
    print("  image_path        Path to the image file to display (default: controls.png)")
    print("  duration_seconds  Display duration in seconds (default: 5)")
    print("  scale_x           Target width in pixels (default: image natural width)")
    print("  scale_y           Target height in pixels (default: image natural height)")
    print()
    print("Converter options:")
    print("  create              Use SVG converter mode")
    print("  -c, --config FILE   Config file to read key-value pairs")
    print("  -i, --input FILE    Input SVG file to modify")
    print("  -r, --readme FILE   README.md file to parse button mappings")
    print("  -p, --port FILE     port.json file to read title from (overrides -c title)")
    print()
    print("Exit conditions:")
    print("  - Specified duration elapses")
    print("  - Any keyboard key is pressed")
    print("  - Any game controller button is pressed")
    print("  - Window is closed")
    print()
    print("Examples:")
    print("  love .")
    print("  love . menu.png")
    print("  love . help.jpg 10")
    print("  love . edgar.png 5 640 480  # 5 sec, scaled to 640x480")
    print("  love . /?  # show this help")
    print("  love . create -c mygame.gptk -i template.svg -r README.md")
    print("  love . create -i template.svg -r README.md -p port.json")
end

local function getActualArgs()
    local cliArgs = {}
    if arg then
        local startIndex = 1
        if #arg >= 1 and arg[1]:sub(1, 1) ~= "-" then
            startIndex = 2
        end

        for i = startIndex, #arg do
            if arg[i] == "--" then
                for j = i + 1, #arg do
                    if arg[j] ~= nil and arg[j] ~= "" then
                        table.insert(cliArgs, arg[j])
                    end
                end
                break
            elseif arg[i] ~= nil and arg[i] ~= "" then
                table.insert(cliArgs, arg[i])
            end
        end
    end
    return cliArgs
end

local function parseArgs(cliArgs)
    if #cliArgs >= 1 and (cliArgs[1] == "/?" or cliArgs[1] == "--help" or cliArgs[1] == "-h") then
        showHelp()
        love.event.quit()
        return
    end

    if #cliArgs >= 1 then
        imagePath = cliArgs[1]
    end

    if #cliArgs >= 2 then
        local durationSeconds = tonumber(cliArgs[2])
        if durationSeconds and durationSeconds > 0 then
            displayDuration = durationSeconds
        end
    end

    if #cliArgs >= 3 then
        targetW = tonumber(cliArgs[3])
    end

    if #cliArgs >= 4 then
        targetH = tonumber(cliArgs[4])
    end
end

local function updateImagePosition()
    windowWidth, windowHeight = love.graphics.getDimensions()
    if controlsImage then
        imageX = (windowWidth - controlsImage:getWidth() * scaleX) / 2
        imageY = (windowHeight - controlsImage:getHeight() * scaleY) / 2
    end
end

function love.load()
    local cliArgs = getActualArgs()
    if #cliArgs >= 1 and (cliArgs[1] == "create" or cliArgs[1] == "convert" or cliArgs[1] == "svg") then
        local ok = converter.run(cliArgs)
        if not ok then
            love.event.quit()
            return
        end
        love.event.quit()
        return
    end

    parseArgs(cliArgs)
    love.window.setFullscreen(true, "desktop")
    love.window.setTitle("Controls")

    local f = io.open(imagePath, "rb")
    if not f then
        print("Image loading failed: Could not open file " .. imagePath .. ". Does not exist.")
        love.event.quit()
        return
    end
    local data = f:read("*a")
    f:close()

    local ok, imageOrError = pcall(function()
        local fileData = love.filesystem.newFileData(data, imagePath)
        return love.graphics.newImage(love.image.newImageData(fileData))
    end)
    if not ok then
        print("Image loading failed:", imageOrError)
        love.event.quit()
        return
    end

    controlsImage = imageOrError
    countdownFont = love.graphics.newFont(48)
    fontHeight = countdownFont:getHeight()
    if targetW then
        scaleX = targetW / controlsImage:getWidth()
    end
    if targetH then
        scaleY = targetH / controlsImage:getHeight()
    end
    updateImagePosition()
end

function love.resize(w, h)
    updateImagePosition()
end

function love.keypressed(key)
    love.event.quit()
end

function love.joystickpressed(joystick, button)
    love.event.quit()
end

function love.update(dt)
    elapsed = elapsed + dt
    if elapsed >= displayDuration then
        love.event.quit()
    end
end

function love.draw()
    love.graphics.clear(0, 0, 0)
    if controlsImage then
        love.graphics.draw(controlsImage, imageX, imageY, 0, scaleX, scaleY)

        local secondsRemaining = math.max(0, math.ceil(displayDuration - elapsed))
        if secondsRemaining > 0 then
            local countdownText = "Starting in " .. secondsRemaining
            local textWidth = countdownFont:getWidth(countdownText)
            love.graphics.setFont(countdownFont)
            love.graphics.setColor(1, 1, 1)
            love.graphics.print(countdownText, (windowWidth - textWidth) / 2, windowHeight - fontHeight - 50)
        end
    end
end
