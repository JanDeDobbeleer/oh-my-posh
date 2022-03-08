-- Duration functions

local endedit_time
local last_duration
local tip_word
local cached_prompt = {}

local function omp_exe()
    return [["::OMP::"]]
end

local function omp_config()
    return [["::CONFIG::"]]
end

local function can_async()
    if (clink.version_encoded or 0) >= 10030001 then
        return settings.get("prompt.async")
    end
end

local function run_posh_command(command)
    command = '"'..command..'"'
    local _,ismain = coroutine.running()
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
        local prompt_exe = string.format('%s --millis', omp_exe())
        return run_posh_command(prompt_exe)
    end
end

local function duration_onbeginedit()
    last_duration = 0
    if endedit_time then
        local beginedit_time = os_clock_millis()
        local elapsed = beginedit_time - endedit_time
        if elapsed >= 0 then
            last_duration = elapsed
        end
    end
end

local function duration_onendedit()
    endedit_time = os_clock_millis()
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
    local prompt_exe = string.format('%s --shell=cmd --config=%s %s %s --rprompt=%s', omp_exe(), omp_config(), execution_time_option(), error_level_option(), rprompt)
    return run_posh_command(prompt_exe)
end

local function get_posh_tooltip(command)
    local prompt_exe = string.format('%s --shell=cmd --config=%s --command="%s"', omp_exe(), omp_config(), command)
    local tooltip = run_posh_command(prompt_exe)
    if tooltip == "" then
        -- If no tooltip, generate normal rprompt.
        tooltip = get_posh_prompt(true)
    end
    return tooltip
end

local p = clink.promptfilter(1)
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
    if tip_word == nil then
        -- No tooltip needed, so generate prompt normally.
        cached_prompt.right = get_posh_prompt(true)
    elseif cached_prompt.tip_space and can_async() then
        -- Generate tooltip asynchronously in response to Spacebar.
        if cached_prompt.coroutine then
            -- Coroutine is already in progress.  The cached right prompt will
            -- be used until the coroutine finishes.
        else
            -- Create coroutine to generate tooltip rprompt.
            cached_prompt.coroutine = coroutine.create(function ()
                cached_prompt.right = get_posh_tooltip(tip_word)
                cached_prompt.tip_done = true
                -- Refresh the prompt once the tooltip is generated.
                clink.refilterprompt()
            end)
        end
        if cached_prompt.tip_done then
            -- Once the tooltip is ready, clear the Spacebar flag so that if the
            -- command word changes and the Spacebar is pressed again, we can
            -- generate a new tooltip.
            cached_prompt.tip_done = nil
            cached_prompt.tip_space = nil
            cached_prompt.coroutine = nil
        end
    else
        -- Tooltip is needed, but not in response to Spacebar, so refresh it
        -- immediately.
        cached_prompt.right = get_posh_tooltip(tip_word)
    end
    return cached_prompt.right, false
end
function p:transientfilter(prompt)
    local prompt_exe = string.format('%s --shell=cmd --config=%s --print-transient', omp_exe(), omp_config())
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

local function builtin_modules_onendedit()
    duration_onendedit()
end

if clink.onbeginedit ~= nil and clink.onendedit ~= nil then
    clink.onbeginedit(builtin_modules_onbeginedit)
    clink.onendedit(builtin_modules_onendedit)
end

-- Tooltips

function ohmyposh_space(rl_buffer)
    rl_buffer:insert(" ")
    local words = string.explode(rl_buffer:getbuffer(), ' ', [["]])
    if words[1] ~= tip_word then
        tip_word = words[1] -- remember the first word for use when filtering the prompt
        cached_prompt.tip_space = can_async()
        clink.refilterprompt() -- invoke the prompt filters so omp can update the prompt per the tip word
    end
end

if rl.setbinding then
    clink.onbeginedit(function () tip_word = nil cached_prompt = {} end)
    rl.setbinding(' ', [["luafunc:ohmyposh_space"]], 'emacs')
end
