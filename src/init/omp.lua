-- Duration functions

local endedit_time
local last_duration

local function os_clock_millis()
    -- Clink v1.2.30 has a fix for Lua's os.clock() implementation failing after
    -- the program has been running more than 24 days.  In older versions, call
    -- OMP to get the time in milliseconds.
    if (clink.version_encoded or 0) >= 10020030 then
        return math.floor(os.clock() * 1000)
    else
        return io.popen("::OMP:: --millis"):read("*n")
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
    local prompt_exe = string.format('::OMP:: --config="::CONFIG::" %s %s --rprompt=%s', execution_time_option(), error_level_option(), rprompt)
    prompt = io.popen(prompt_exe):read("*a")
    return prompt
end

local p = clink.promptfilter(1)
function p:filter(prompt)
    return get_posh_prompt(false)
end
function p:rightfilter(prompt)
    return get_posh_prompt(true), false
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
