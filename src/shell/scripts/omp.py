import uuid

$POWERLINE_COMMAND = "oh-my-posh"
$POSH_THEME = r"::CONFIG::"
$POSH_PID = uuid.uuid4().hex
$POSH_SHELL_VERSION = $XONSH_VERSION

def get_command_context():
    last_cmd = __xonsh__.history[-1] if __xonsh__.history else None
    status = last_cmd.rtn if last_cmd else 0
    duration = round((last_cmd.ts[1] - last_cmd.ts[0]) * 1000) if last_cmd else 0
    return status, duration

def posh_primary():
    status, duration = get_command_context()
    return $(::OMP:: print primary --config=@($POSH_THEME) --shell=xonsh --status=@(status) --execution-time=@(duration) --shell-version=@($POSH_SHELL_VERSION))

def posh_right():
    status, duration = get_command_context()
    return $(::OMP:: print right --config=@($POSH_THEME) --shell=xonsh --status=@(status) --execution-time=@(duration) --shell-version=@($POSH_SHELL_VERSION))


$PROMPT = posh_primary
$RIGHT_PROMPT = posh_right

notice = """::UPGRADENOTICE::"""

if "::UPGRADE::" == "true":
  print(notice)

if "::AUTOUPGRADE::" == "true":
  ::OMP:: upgrade
