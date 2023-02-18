import uuid

$POWERLINE_COMMAND = "oh-my-posh"
$POSH_THEME = "::CONFIG::"
$POSH_PID = uuid.uuid4().hex

def get_command_context():
    last_cmd = __xonsh__.history[-1] if __xonsh__.history else None
    status = last_cmd.rtn if last_cmd else 0
    duration = round((last_cmd.ts[1] - last_cmd.ts[0]) * 1000) if last_cmd else 0
    return status, duration

def posh_primary():
    status, duration = get_command_context()
    return $(::OMP:: print primary --config=@($POSH_THEME) --shell=xonsh --error=@(status) --execution-time=@(duration) | cat)

def posh_right():
    status, duration = get_command_context()
    return $(::OMP:: print right --config=@($POSH_THEME) --shell=xonsh --error=@(status) --execution-time=@(duration) | cat)


$PROMPT = posh_primary
$RIGHT_PROMPT = posh_right
