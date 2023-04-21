-- Upgrade notice

local notice = [[::UPGRADENOTICE::]]

if '::UPGRADE::' == 'true' then
    print(notice)
end

-- Helper functions

local function get_priority_number(name, default)
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

-- Environment variables

local function environment_onbeginedit()
    os.setenv("POSH_CURSOR_LINE", console.getnumlines())
end

-- Local state

local endedit_time = 0
local last_duration = 0
local tooltips_enabled = ::TOOLTIPS::
local rprompt_enabled = ::RPROMPT::

local cached_prompt = {}
-- Fields in cached_prompt:
--      .cwd            = Current working directory of prompt.
--      .left           = Left side prompt.
--      .right          = Right side prompt.
--      .tooltip        = Tooltip prompt.
--      .tip_command    = Command for which to produce a tooltip.
--      .coroutine      = Coroutine for the tooltip prompt.

local function cache_onbeginedit()
    local cwd = os.getcwd()
    local old_cache = cached_prompt

    -- Start a new table for the new edit/prompt session.
    cached_prompt = { cwd=cwd }

    -- Copy the cached left/right prompt strings if the cwd hasn't changed.
    -- IMPORTANT OPTIMIZATION:  This keeps the prompt highly responsive, except
    -- when changing the current working directory.
    if old_cache.cwd == cwd then
        cached_prompt.left = old_cache.left
        cached_prompt.right = old_cache.right
    end
end

-- Configuration

local function omp_exe()
    return '"'..::OMP::..'"'
end

local function omp_config()
    return '"'..::CONFIG::..'"'
end

os.setenv("POSH_THEME", ::CONFIG::)
os.setenv("POSH_SHELL_VERSION", string.format('clink v%s.%s.%s.%s', clink.version_major, clink.version_minor, clink.version_patch, clink.version_commit))

-- Execution helpers

local function can_async()
    if (clink.version_encoded or 0) >= 10030001 then
        return settings.get("prompt.async")
    end
end

local function run_posh_command(command)
    command = '"'..command..'"'
    local _, ismain = coroutine.running()
    local output
    if ismain then
        output = io.popen(command):read("*a")
    else
        output = io.popenyield(command):read("*a")
    end
    return output
end

-- Duration functions

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

local function set_posh_tooltip(tip_command)
    if tip_command ~= "" and tip_command ~= cached_prompt.tip_command then
        -- Escape special characters properly, if any.
        local escaped_tip_command = string.gsub(tip_command, '(\\+)"', '%1%1"'):gsub('(\\+)$', '%1%1'):gsub('"', '\\"'):gsub('([&<>%(%)@%^|])', '^%1')

        local prompt_exe = string.format('%s print tooltip --shell=cmd %s --config=%s --command="%s"', omp_exe(), error_level_option(), omp_config(), escaped_tip_command)
        local tooltip = run_posh_command(prompt_exe)
        -- Do not cache an empty tooltip.
        if tooltip == "" then
            return
        end
        cached_prompt.tip_command = tip_command
        cached_prompt.tooltip = tooltip
    end
end

local function display_cached_prompt()
    -- Use what's already cached; avoid running oh-my-posh.
    cached_prompt.only_use_cache = true
    clink.refilterprompt()
    cached_prompt.only_use_cache = nil
end

local function async_collect_posh_prompts()
    -- Generate the left prompt.
    cached_prompt.left = get_posh_prompt(false)

    -- Generate the right prompt, if needed.
    if rprompt_enabled then
        display_cached_prompt() -- Show left side; don't wait for right side.
        cached_prompt.right = get_posh_prompt(true)
    end
end

-- set priority lower than z.lua
-- https://github.com/skywind3000/z.lua/pull/125/commits/48a77adf3575952b2e951aa820a1ce11ed4ce56b
local zl_prompt_priority = get_priority_number('_ZL_CLINK_PROMPT_PRIORITY', 0)
local p = clink.promptfilter(zl_prompt_priority + 1)
function p:filter(prompt)
    local need_left = true

    -- Get a left prompt immediately if nothing is available yet.
    if not cached_prompt.left then
        cached_prompt.left = get_posh_prompt(false)
        need_left = false
    end

    -- Get left/right prompts asynchronously, if possible.
    if not cached_prompt.only_use_cache then
        if can_async() then
            -- IMPORTANT:  Defining this function inline makes sure it only
            -- updates the same cached_prompt table that existed when the
            -- function was defined.  That way if a new prompt starts (which
            -- discards the old coroutine) and a new coroutine starts, the old
            -- coroutine won't stomp on the new cached_prompt table.
            clink.promptcoroutine(function ()
                -- Generate left prompt, if needed.
                if need_left then
                    cached_prompt.left = get_posh_prompt(false)
                end
                -- Generate right prompt, if needed.
                if rprompt_enabled then
                    if need_left then
                        -- Show left side while right side is being generated.
                        display_cached_prompt()
                    end
                    cached_prompt.right = get_posh_prompt(true)
                else
                    cached_prompt.right = nil
                end
            end)
        else
            if need_left then
                cached_prompt.left = get_posh_prompt(false)
            end
            if rprompt_enabled then
                cached_prompt.right = get_posh_prompt(true)
            end
        end
    end

    return cached_prompt.left
end
function p:rightfilter(prompt)
    -- Return cached tooltip if available, otherwise return cached rprompt.
    -- Returning false as the second return value halts further prompt
    -- filtering, to keep other things from overriding what we generated.
    return (cached_prompt.tooltip or cached_prompt.right), false
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
    cache_onbeginedit()
    duration_onbeginedit()
    environment_onbeginedit()
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
    -- Insert space first, in case it might affect the tip word, e.g. it could
    -- split "gitcommit" into "git commit".
    rl_buffer:insert(" ")
    -- Get the first word of command line as tip.
    local tip_command = rl_buffer:getbuffer():gsub("^%s*([^%s]*).*$", "%1")

    -- Generate a tooltip asynchronously (via coroutine) if available, otherwise
    -- generate a tooltip immediately.
    if not can_async() then
        set_posh_tooltip(tip_command)
        clink.refilterprompt()
    elseif cached_prompt.coroutine then
        -- No action needed; a tooltip coroutine is already running.
    else
        cached_prompt.coroutine = coroutine.create(function ()
            set_posh_tooltip(tip_command)
            if cached_prompt.coroutine == coroutine.running() then
                cached_prompt.coroutine = nil
            end
            display_cached_prompt()
        end)
    end
end

if tooltips_enabled and rl.setbinding then
    rl.setbinding(' ', [["luafunc:ohmyposh_space"]], 'emacs')
end
