-- Helper functions

function get_priority_number(name, default)
	local value = os.getenv(name)
	if os.envmap ~= nil and type(os.envmap) == 'table' then
		local t = os.envmap[name]
		value = (t ~= nil and type(t) == 'string') and t or value
	end
	if type(default) == 'number' then
		value = tonumber(value)
		if value == nil then
			return default
		else
			return value
		end
	else
        return default
	end
end

-- Duration functions

local endedit_time = 0
local last_duration = 0
local tip
local tooltips_enabled = ::TOOLTIPS::
local tooltip_active = false
local cached_prompt = {}

local function omp_exe()
    return '"'..::OMP::..'"'
end

local function omp_config()
    return '"'..::CONFIG::..'"'
end

os.setenv("POSH_THEME", ::CONFIG::)

local function can_async()
    if (clink.version_encoded or 0) >= 10030001 then
        return settings.get("prompt.async")
    end
end

local function run_posh_command(command)
    command = '"'..command..'"'
    local _,ismain = coroutine.running()
    local output
    if ismain then
        output = io.popen(command):read("*a")
    else
        output = io.popenyield(command):read("*a")
    end
    return output
end

local function os_clock_millis()
    -- Clink v1.2.30 has a fix for Lua's os.clock() implementation failing after
    -- the program has been running more than 24 days.  In older versions, call
    -- OMP to get the time in milliseconds.
    if (clink.version_encoded or 0) >= 10020030 then
        return math.floor(os.clock() * 1000)
    else
        local prompt_exe = string.format('%s get millis --shell=cmd', omp_exe())
        return run_posh_command(prompt_exe)
    end
end

local function duration_onbeginedit()
    last_duration = 0
    if endedit_time ~= 0 then
        local beginedit_time = os_clock_millis()
        local elapsed = beginedit_time - endedit_time
        if elapsed >= 0 then
            last_duration = elapsed
        end
    end
end

local function duration_onendedit(input)
    endedit_time = 0
    -- For an empty command, the execution time should not be evaluated.
    if string.gsub(input, "^%s*(.-)%s*$", "%1") ~= "" then
        endedit_time = os_clock_millis()
    end
end

-- Prompt functions

local function execution_time_option()
    if last_duration ~= nil then
        return "--execution-time "..last_duration
    end
    return ""
end

local function error_level_option()
    if os.geterrorlevel ~= nil and settings.get("cmd.get_errorlevel") then
        return "--error "..os.geterrorlevel()
    end
    return ""
end

local function get_posh_prompt(rprompt)
    local prompt = "primary"
    if rprompt then
        prompt = "right"
    end
    local prompt_exe = string.format('%s print %s --shell=cmd --config=%s %s %s', omp_exe(), prompt, omp_config(), execution_time_option(), error_level_option(), rprompt)
    return run_posh_command(prompt_exe)
end

local function set_posh_tooltip(command)
    if command == nil then
        return
    end

    -- escape special characters properly, if any
    command = string.gsub(command, '(\\+)"', '%1%1"')
    command = string.gsub(command, '(\\+)$', '%1%1')
    command = string.gsub(command, '"', '\\"')
    command = string.gsub(command, '([&<>%(%)@%^|])', '^%1')

    local prompt_exe = string.format('%s print tooltip --shell=cmd %s --config=%s --command="%s"', omp_exe(), error_level_option(), omp_config(), command)
    local tooltip = run_posh_command(prompt_exe)
    if tooltip ~= "" then
        tooltip_active = true
        cached_prompt.right = tooltip
    end
end

-- set priority lower than z.lua
-- https://github.com/skywind3000/z.lua/pull/125/commits/48a77adf3575952b2e951aa820a1ce11ed4ce56b
local zl_prompt_priority = get_priority_number('_ZL_CLINK_PROMPT_PRIORITY', 0)
local p = clink.promptfilter(zl_prompt_priority + 1)
function p:filter(prompt)
    if cached_prompt.left and cached_prompt.tip_space then
        -- Use the cached left prompt when updating the rprompt (tooltip) in
        -- response to the Spacebar.  This allows typing to stay responsive.
    else
        -- Generate the left prompt normally.
        cached_prompt.left = get_posh_prompt(false)
    end
    return cached_prompt.left
end
function p:rightfilter(prompt)
    if cached_prompt.tip_space and can_async() then
        -- Generate tooltip asynchronously in response to Spacebar.
        if cached_prompt.coroutine then
            -- Coroutine is already in progress.  The cached right prompt will
            -- be used until the coroutine finishes.
        else
            -- Create coroutine to generate tooltip rprompt.
            cached_prompt.coroutine = coroutine.create(function ()
                set_posh_tooltip(tip)
                cached_prompt.tip_done = true
                -- Refresh the prompt once the tooltip is generated.
                clink.refilterprompt()
            end)
        end
        if cached_prompt.tip_done then
            -- Once the tooltip is ready, clear the Spacebar flag so that if the
            -- tip changes and the Spacebar is pressed again, we can
            -- generate a new tooltip.
            cached_prompt.tip_done = nil
            cached_prompt.tip_space = nil
            cached_prompt.coroutine = nil
        end
    else
        -- Tooltip is needed, but not in response to Spacebar, so refresh it
        -- immediately.
        set_posh_tooltip(tip)
    end
    if not tooltip_active then
        -- Tooltip is not active, generate rprompt normally.
        cached_prompt.right = get_posh_prompt(true)
    end
    return cached_prompt.right, false
end
function p:transientfilter(prompt)
    local prompt_exe = string.format('%s print transient --shell=cmd --config=%s %s', omp_exe(), omp_config(), error_level_option())
    prompt = run_posh_command(prompt_exe)
    if prompt == "" then
        prompt = nil
    end
    return prompt
end
function p:transientrightfilter(prompt)
    return "", false
end

-- Event handlers

local function builtin_modules_onbeginedit()
    _cached_state = {}
    duration_onbeginedit()
end

local function builtin_modules_onendedit(input)
    duration_onendedit(input)
end

if clink.onbeginedit ~= nil and clink.onendedit ~= nil then
    clink.onbeginedit(builtin_modules_onbeginedit)
    clink.onendedit(builtin_modules_onendedit)
end

-- Tooltips

function ohmyposh_space(rl_buffer)
    local new_tip = string.gsub(rl_buffer:getbuffer(), "^%s*(.-)%s*$", "%1")
    rl_buffer:insert(" ")
    if new_tip ~= tip then
        tip = new_tip -- remember the tip for use when filtering the prompt
        cached_prompt.tip_space = can_async()
        clink.refilterprompt() -- invoke the prompt filters so OMP can update the prompt per the tip
    end
end

if tooltips_enabled and rl.setbinding then
    clink.onbeginedit(function () tip = nil cached_prompt = {} end)
    rl.setbinding(' ', [["luafunc:ohmyposh_space"]], 'emacs')
end
