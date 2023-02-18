set-env POSH_PID (to-string (randint 10000000000000 10000000000000000))
set-env POSH_THEME '::CONFIG::'
set-env POWERLINE_COMMAND 'oh-my-posh'
set-env ELVISH_VERSION (elvish --version)

var error-code = 0

fn posh-after-command-hook {|m|
    var error = $m[error]
    if (is $error $nil) {
        set error-code = 0
    } else {
        try {
            set error-code = $error[reason][exit-status]
        } catch {
            # built-in commands don't have a status code.
            set error-code = 1
        }
    }
}

set edit:after-command = [ $@edit:after-command $posh-after-command-hook~ ]

set edit:prompt = {
    var cmd-duration = (printf "%.0f" (* $edit:command-duration 1000))
    ::OMP:: print primary --config=$E:POSH_THEME --shell=elvish --execution-time=$cmd-duration --error=$error-code --pwd=$pwd --shell-version=$E:ELVISH_VERSION
}

set edit:rprompt = {
    var cmd-duration = (printf "%.0f" (* $edit:command-duration 1000))
    ::OMP:: print right --config=$E:POSH_THEME --shell=elvish --execution-time=$cmd-duration --error=$error-code --pwd=$pwd --shell-version=$E:ELVISH_VERSION
}
