set-env POSH_PID (to-string (randint 10000000000000 10000000000000000))
set-env POSH_THEME ::CONFIG::
set-env POSH_SHELL_VERSION (elvish --version)
set-env POWERLINE_COMMAND 'oh-my-posh'

var _omp_error_code = 0
var _omp_executable = ::OMP::

fn posh-after-command-hook {|m|
    var error = $m[error]
    if (is $error $nil) {
        set _omp_error_code = 0
    } else {
        try {
            set _omp_error_code = $error[reason][exit-status]
        } catch {
            # built-in commands don't have a status code.
            set _omp_error_code = 1
        }
    }
}

set edit:after-command = [ $@edit:after-command $posh-after-command-hook~ ]

set edit:prompt = {
    var cmd-duration = (printf "%.0f" (* $edit:command-duration 1000))
    (external $_omp_executable) print primary --shell=elvish --execution-time=$cmd-duration --status=$_omp_error_code --pwd=$pwd --shell-version=$E:POSH_SHELL_VERSION
}

set edit:rprompt = {
    var cmd-duration = (printf "%.0f" (* $edit:command-duration 1000))
    (external $_omp_executable) print right --shell=elvish --execution-time=$cmd-duration --status=$_omp_error_code --pwd=$pwd --shell-version=$E:POSH_SHELL_VERSION
}
