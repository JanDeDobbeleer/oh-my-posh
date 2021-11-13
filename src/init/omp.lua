-- Duration functions

local endedit_time
local last_duration

local function duration_onbeginedit()
    last_duration = 0
    if endedit_time then
        local beginedit_time = io.popen("::OMP:: --millis"):read("*n")
        local elapsed = beginedit_time - endedit_time
        if elapsed >= 0 then
            last_duration = elapsed
        end
    end
end

local function duration_onendedit()
    endedit_time = io.popen("::OMP:: --millis"):read("*n")
end

-- Prompt functions

local function get_posh_prompt(rprompt)
    local prompt_exe = string.format("::OMP:: --config=\"::CONFIG::\" --execution-time %s --error %s --rprompt=%s", last_duration, os.geterrorlevel(), rprompt)
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


clink.onbeginedit(builtin_modules_onbeginedit)
clink.onendedit(builtin_modules_onendedit)
