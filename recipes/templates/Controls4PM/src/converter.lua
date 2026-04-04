local M = {}

local function trim(s)
    if not s then
        return s
    end
    return s:match("^%s*(.-)%s*$")
end

local function escape_pattern(text)
    return text:gsub("([%%%^%$%(%)%.%[%]%*%+%-%?])", "%%%1")
end

local function escape_replacement(text)
    return text:gsub("%%", "%%%%")
end

local function wrap_text(text, max_chars)
    if #text <= max_chars then
        return {text}
    end
    local lines = {}
    local remaining = text
    while #remaining > max_chars do
        local break_pos = nil
        for i = max_chars, 1, -1 do
            if remaining:sub(i, i) == " " then
                break_pos = i
                break
            end
        end
        if break_pos then
            table.insert(lines, remaining:sub(1, break_pos - 1))
            remaining = remaining:sub(break_pos + 1)
        else
            table.insert(lines, remaining:sub(1, max_chars))
            remaining = remaining:sub(max_chars + 1)
        end
    end
    if #remaining > 0 then
        table.insert(lines, remaining)
    end
    return lines
end

local function find_group_end(content, start)
    local pos = start
    local depth = 1
    while depth > 0 and pos <= #content do
        local open = content:find("<g[%s>]", pos)
        local close = content:find("</g>", pos)
        if not close then break end
        if open and open < close then
            depth = depth + 1
            pos = open + 2
        else
            depth = depth - 1
            if depth == 0 then
                return close + 3
            end
            pos = close + 4
        end
    end
    return nil
end

local function wrap_long_text_in_svg(content, max_chars)
    local summary_start = content:find('<g[^>]*id="summary"')
    local summary_end = nil
    if summary_start then
        local group_tag_end = content:find(">", summary_start)
        if group_tag_end then
            summary_end = find_group_end(content, group_tag_end + 1)
        end
    end

    local parts = {}
    local pos = 1

    while pos <= #content do
        local ts = content:find("<tspan", pos, true)
        if not ts then
            table.insert(parts, content:sub(pos))
            break
        end

        table.insert(parts, content:sub(pos, ts - 1))

        local te = content:find(">", ts)
        if not te then
            table.insert(parts, content:sub(ts))
            break
        end

        local open_tag = content:sub(ts, te)

        -- Self-closing tspan (<tspan ... />) — pass through unchanged
        if open_tag:sub(-2) == "/>" then
            table.insert(parts, open_tag)
            pos = te + 1
        else
            local tc = content:find("</tspan>", te + 1, true)
            if not tc then
                table.insert(parts, content:sub(ts))
                break
            end

            local tspan_text = content:sub(te + 1, tc - 1)
            local in_summary = summary_start and summary_end and
                               ts >= summary_start and ts <= summary_end

            if not in_summary and #tspan_text > max_chars then
                local lines = wrap_text(tspan_text, max_chars)
                local x_val = open_tag:match('x="([^"]*)"') or ""
                local y_val = tonumber(open_tag:match('y="([^"]*)"')) or 0
                table.insert(parts, open_tag .. lines[1] .. "</tspan>")
                for i = 2, #lines do
                    local line_y = y_val + (i - 1) * 26
                    table.insert(parts, '\n          <tspan sodipodi:role="line" x="' .. x_val .. '" y="' .. line_y .. '">' .. lines[i] .. "</tspan>")
                end
                local spacer_y = y_val + #lines * 26
                table.insert(parts, '\n          <tspan sodipodi:role="line" x="' .. x_val .. '" y="' .. spacer_y .. '"></tspan>')
            else
                table.insert(parts, open_tag .. tspan_text .. "</tspan>")
            end

            pos = tc + 8
        end
    end

    return table.concat(parts)
end

local function read_file(path)
    local f, err = io.open(path, "r")
    if not f then
        return nil, err
    end
    local data = f:read("*a")
    f:close()
    return data
end

local function replace_all_literal(content, placeholder, replacement)
    local p = content:find(placeholder, 1, true)
    while p do
        content = content:sub(1, p - 1) .. replacement .. content:sub(p + #placeholder)
        p = content:find(placeholder, 1, true)
    end
    return content
end

local function read_port_title(json_path)
    local content, err = read_file(json_path)
    if not content then
        return nil, err
    end
    local attr_block = content:match('"attr"%s*:%s*(%b{})')
    if not attr_block then
        return nil, "No 'attr' section found in " .. json_path
    end
    local title = attr_block:match('"title"%s*:%s*"([^"]*)"')
    if not title then
        return nil, "No 'title' found in attr section of " .. json_path
    end
    return title
end

local function split_cli_args(arg_table)
    local args = {}
    if not arg_table then
        return args
    end

    local startIndex = 1
    if #arg_table >= 1 and (arg_table[1] == "." or arg_table[1] == "./" or arg_table[1] == ".." or arg_table[1]:sub(-5) == ".love") then
        startIndex = 2
    end

    for i = startIndex, #arg_table do
        if arg_table[i] == "--" then
            for j = i + 1, #arg_table do
                table.insert(args, arg_table[j])
            end
            break
        elseif arg_table[i] and arg_table[i] ~= "" then
            table.insert(args, arg_table[i])
        end
    end

    return args
end

local function show_help()
    print("Usage: love . create -c CONFIG_FILE -i INPUT_SVG -r README_FILE")
    print()
    print("Options:")
    print("  -c, --config FILE    Config file to read key-value pairs")
    print("  -i, --input FILE     Input SVG file to modify")
    print("  -r, --readme FILE    README.md file to parse button mappings")
    print("  -p, --port FILE      port.json file to read title from (overrides -c title)")
    print("  -h, --help           Show this help message")
    print()
    print("Examples:")
    print("  love . create -c mygame.gptk -i template.svg -r /home/user/port/README.md")
    print("  love . create --config controls.cfg --input input.svg --readme /home/user/port/README.md")
end

local function parse_args(args)
    local parsed = {
        config_file = nil,
        svg_file = nil,
        readme_file = nil,
        port_file = nil,
        show_help = false,
        error = nil,
    }

    local i = 1
    while i <= #args do
        local token = args[i]
        if token == "-c" or token == "--config" then
            i = i + 1
            parsed.config_file = args[i]
        elseif token == "-i" or token == "--input" then
            i = i + 1
            parsed.svg_file = args[i]
        elseif token == "-r" or token == "--readme" then
            i = i + 1
            parsed.readme_file = args[i]
        elseif token == "-p" or token == "--port" then
            i = i + 1
            parsed.port_file = args[i]
        elseif token == "-h" or token == "--help" then
            parsed.show_help = true
        else
            parsed.error = "Unknown option: " .. token
            break
        end
        i = i + 1
    end

    if not parsed.show_help and not parsed.error then
        if not parsed.svg_file then
            parsed.error = "Missing --input option"
        elseif not parsed.readme_file then
            parsed.error = "Missing --readme option"
        end
    end

    return parsed
end

local function extract_buttons_from_file(readme_path)
    local button_mapping = {
        A = "n/a",
        B = "n/a",
        X = "n/a",
        Y = "n/a",
        L1 = "n/a",
        L2 = "n/a",
        R1 = "n/a",
        R2 = "n/a",
        SELECT = "n/a",
        START = "n/a",
    }

    local handle = io.open(readme_path, "r")
    if not handle then
        return nil, "Could not open README file: " .. readme_path
    end

    for line in handle:lines() do
        local button, action = line:match("^|%s*([A-Za-z0-9]+)%s*|%s*([^|]+)%s*|")
        if button and action then
            button = button:upper()
            if button_mapping[button] then
                button_mapping[button] = trim(action)
            end
        end
    end
    handle:close()

    return button_mapping
end

local function replace_placeholders(content, button_mapping)
    local buttons = {"A", "B", "X", "Y", "L1", "L2", "R1", "R2"}
    for _, button in ipairs(buttons) do
        local placeholder
        if #button == 1 then
            placeholder = "Z" .. button .. "D1"
        elseif #button == 2 then
            placeholder = "Z" .. button:sub(1, 1) .. "D" .. button:sub(2, 2)
        else
            placeholder = "Z" .. button
        end

        local replacement = button_mapping[button] or ""
        local escapedReplacement = escape_replacement(replacement)
        content = content:gsub(escape_pattern(placeholder), escapedReplacement)
    end

    return content
end

local function apply_config_replacements(content, config_path, button_mapping)
    local handle = io.open(config_path, "r")
    if not handle then
        return nil, "Could not open config file: " .. config_path
    end

    local select_gptk = nil
    local start_gptk = nil

    for line in handle:lines() do
        local key, value = line:match("^%s*([^=]+)%s*=%s*(.-)%s*$")
        if key and value and key ~= "" and value ~= "" then
            key = trim(key)
            value = trim(value)
            local zkey = key:upper()

            if key == "select" or key == "back" then
                select_gptk = value
            elseif key == "start" then
                start_gptk = value
            elseif key == "a" or key == "b" or key == "x" or key == "y" or
                   key == "up" or key == "down" or key == "left" or key == "right" or
                   key == "l1" or key == "l2" or key == "r1" or key == "r2" then
                local escapedReplacement = escape_replacement(zkey)
                content = content:gsub(escape_pattern("Z" .. zkey), escapedReplacement)
            end
        end
    end

    -- README takes priority; gptk is fallback
    local select_value
    if button_mapping["SELECT"] ~= "n/a" then
        select_value = button_mapping["SELECT"]
    elseif select_gptk and select_gptk ~= "" then
        select_value = select_gptk
    else
        select_value = "Escape"
    end

    local start_value
    if button_mapping["START"] ~= "n/a" then
        start_value = button_mapping["START"]
    elseif start_gptk and start_gptk ~= "" then
        start_value = start_gptk
    else
        start_value = "Enter"
    end

    local escapedSelect = escape_replacement(select_value)
    local escapedStart = escape_replacement(start_value)
    content = content:gsub(escape_pattern("ZSELECT"), escapedSelect)
    content = content:gsub(escape_pattern("ZSTART"), escapedStart)

    handle:close()
    return content
end

function M.run(arg_table)
    local args = split_cli_args(arg_table)
    if #args >= 1 and (args[1] == "create" or args[1] == "convert" or args[1] == "svg") then
        table.remove(args, 1)
    end

    local parsed = parse_args(args)
    if parsed.show_help then
        show_help()
        return true
    end
    if parsed.error then
        print(parsed.error)
        show_help()
        return false
    end

    local config_basename
    if parsed.config_file then
        config_basename = parsed.config_file:gsub("%.gptk$", "")
        config_basename = config_basename:match("([^/\\]+)$") or config_basename
    else
        config_basename = parsed.svg_file:gsub("%.[^.]+$", "")
        config_basename = config_basename:match("([^/\\]+)$") or config_basename
    end
    local title = config_basename:gsub("^%l", string.upper)

    if parsed.port_file then
        local port_title, port_err = read_port_title(parsed.port_file)
        if port_title then
            title = port_title
        else
            print("Warning: " .. (port_err or "unknown error"))
        end
    end

    local svg_content, err = read_file(parsed.svg_file)
    if not svg_content then
        print(err)
        return false
    end

    local button_mapping, mapping_err = extract_buttons_from_file(parsed.readme_file)
    if not button_mapping then
        print(mapping_err)
        return false
    end

    local content = replace_placeholders(svg_content, button_mapping)
    if parsed.config_file then
        content, err = apply_config_replacements(content, parsed.config_file, button_mapping)
        if not content then
            print(err)
            return false
        end
    else
        local select_value = button_mapping["SELECT"] ~= "n/a" and button_mapping["SELECT"] or "Escape"
        local start_value  = button_mapping["START"]  ~= "n/a" and button_mapping["START"]  or "Enter"
        content = replace_all_literal(content, "ZSELECT", select_value)
        content = replace_all_literal(content, "ZSTART", start_value)
    end
    content = replace_all_literal(content, "ZTITLE", title)

    local dpad_defaults = {ZUP = "Up", ZDOWN = "Down", ZLEFT = "Left", ZRIGHT = "Right"}
    for placeholder, default in pairs(dpad_defaults) do
        content = replace_all_literal(content, placeholder, default)
    end

    content = wrap_long_text_in_svg(content, 10)

    local output_file = "controls.svg"
    local out_handle, out_err = io.open(output_file, "w")
    if not out_handle then
        print("Error writing output SVG:", out_err)
        return false
    end
    out_handle:write(content)
    out_handle:close()

    local png_file = "controls.png"
    local inkscape_command = string.format("inkscape %q -o %q", output_file, png_file)
    local status = os.execute(inkscape_command)
    if status ~= 0 then
        print("Warning: inkscape command failed or returned non-zero status")
    end

    print("Replacements complete!")
    print("Modified file saved as: " .. output_file)
    print("PNG file created as: " .. png_file)

    return true
end

return M
