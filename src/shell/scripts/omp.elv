set-env POSH_SESSION_ID ::SESSION_ID::
set-env POSH_THEME ::CONFIG::
set-env POSH_SHELL elvish
set-env POSH_SHELL_VERSION $version
set-env POWERLINE_COMMAND oh-my-posh

# disable all known python virtual environment prompts
set-env VIRTUAL_ENV_DISABLE_PROMPT 1
set-env PYENV_VIRTUALENV_DISABLE_PROMPT 1

var _omp_executable = (external ::OMP::)
var _omp_status = 0
var _omp_no_status = 1
var _omp_execution_time = -1
var _omp_terminal_width = ($_omp_executable get width)

fn _omp-after-readline-hook {|_|
    set _omp_execution_time = -1

    # Getting the terminal width can fail inside a prompt function, so we do this here.
    set _omp_terminal_width = ($_omp_executable get width)
}

fn _omp-after-command-hook {|m|
    # The command execution time should not be available in the first prompt.
    if (== $_omp_no_status 0) {
        set _omp_execution_time = (printf %.0f (* $m[duration] 1000))
    }

    set _omp_no_status = 0

    var error = $m[error]
    if (is $error $nil) {
        set _omp_status = 0
    } else {
        try {
            set _omp_status = $error[reason][exit-status]
        } catch {
            # built-in commands don't have a status code.
            set _omp_status = 1
        }
    }
}

fn _omp_get_prompt {|type @arguments|
    $_omp_executable print $type ^
        --save-cache ^
        --shell=elvish ^
        --shell-version=$E:POSH_SHELL_VERSION ^
        --status=$_omp_status ^
        --no-status=$_omp_no_status ^
        --execution-time=$_omp_execution_time ^
        --terminal-width=$_omp_terminal_width ^
        $@arguments
}

set edit:after-readline = [ $@edit:after-readline $_omp-after-readline-hook~ ]
set edit:after-command = [ $@edit:after-command $_omp-after-command-hook~ ]
set edit:prompt = {|| _omp_get_prompt primary }
set edit:rprompt = {|| _omp_get_prompt right }
