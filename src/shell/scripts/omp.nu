let _omp_executable: string = (echo ::OMP::)

let _omp_executable_is_path = (
    ($_omp_executable | str contains "/")
    or ($_omp_executable | str contains "\\")
    or ($_omp_executable | str starts-with ".")
    or ($_omp_executable | str starts-with "~")
)

if not (($_omp_executable_is_path and ($_omp_executable | path exists)) or (which $_omp_executable | is-not-empty)) { return }

if ($env.config? | is-not-empty) {
    $env.config = ($env.config | upsert render_right_prompt_on_last_line true)
}

$env.POWERLINE_COMMAND = 'oh-my-posh'
$env.PROMPT_INDICATOR = ""
$env._OMP_ENABLE_STREAMING = false
$env._OMP_SERVE_JOB = null
$env._OMP_CYCLE = 0
$env.POSH_SESSION_ID = "::SESSION_ID::"
# pinned to the session's configuration so it can be recovered when the session cache is lost
$env.POSH_CONFIG = ::CONFIG::
$env.POSH_SHELL = "nu"
$env.POSH_SHELL_VERSION = (version | get version)

$env.VIRTUAL_ENV_DISABLE_PROMPT = 1
$env.PYENV_VIRTUALENV_DISABLE_PROMPT = 1

def _omp_context [] {
    # We have to do this because the initial value of `$env.CMD_DURATION_MS` is always `0823`, which is an official setting.
    # See https://github.com/nushell/nushell/discussions/6402#discussioncomment-3466687.
    let execution_time = match $env.CMD_DURATION_MS {
        '0823' => -1
        $ms => { $ms | into int }
    }

    let no_status = if $nu.history-enabled {
        not ($env.POSH_EXECUTED? | default false)
    } else {
        $execution_time < 0
    }

    {
        shell: nu
        shell-version: $env.POSH_SHELL_VERSION
        status: $env.LAST_EXIT_CODE
        no-status: $no_status
        execution-time: $execution_time
        terminal-width: (term size).columns
        job-count: (job list | where id not-in [$env._OMP_SERVE_JOB $env._OMP_CYCLE] | length)
    }
}

def --wrapped _omp_get_prompt [
    type: string,
    ...args: string
] {
    (
        ^$_omp_executable print $type
            --save-cache
            ...(_omp_context | items { |key, value| $"--($key)=($value)" })
            ...$args
    )
}

$env.PROMPT_MULTILINE_INDICATOR = (
    ^$_omp_executable print secondary
        --shell=nu
        $"--shell-version=($env.POSH_SHELL_VERSION)"
)

def --env _omp_start_serve [] {
    $env._OMP_SERVE_JOB = (job spawn --auto-kill {
        (
            generate { |_| {out: (job recv | to json --raw), next: true} } true
            | to text
            | ^$_omp_executable serve --shell=nu --json
            | lines
        ) | reduce --fold {id: 0, prompt: ''} { |line, shown|
            let record = ($line | from json)
            if ($record.prompt | str starts-with (char rs)) {
                $record.prompt | str trim --left --char (char rs) | job send 0 --tag $record.id
                return $shown
            }
            if $record.id != $shown.id {
                $record.prompt | job send 0 --tag $record.id
                return $record
            }
            if $record.prompt != $shown.prompt {
                try { $record.prompt | job send $record.id }
            }
            $record
        } | ignore
    })
}

def _omp_env [] {
    (
        $env
        | transpose key value
        | where ($it.value | describe) in [string int float bool]
        | update value { into string }
        | append {key: PATH, value: ($env.PATH | str join (char esep))}
        | transpose --header-row --as-record
    )
}

def --env _omp_primary [] {
    if ($env.SET_POSHCONTEXT? | is-not-empty) {
        do --env $env.SET_POSHCONTEXT
    }

    mut cleared = false
    if $env._OMP_ENABLE_STREAMING {
        # pre_execution tracks what history would tell us, without reading it on every prompt
        $env.POSH_EXECUTED = $env._OMP_EXECUTED
        $env._OMP_EXECUTED = false
        $cleared = $env._OMP_CLEARED
    } else {
        let hist = if $nu.history-enabled { history } else { [] }
        let hist_len = ($hist | length)

        # a command was executed this prompt cycle only if history actually grew;
        # an empty Enter (or history disabled) leaves the length unchanged
        $env.POSH_EXECUTED = ($nu.history-enabled and ($hist_len > ($env.POSH_LAST_HISTORY_LEN? | default 0)))
        $env.POSH_LAST_HISTORY_LEN = $hist_len

        # hack: sets cursor line to 1 on clear; not bulletproof, just a start
        $cleared = $nu.history-enabled and (($hist | is-empty) or ($hist | last | get command?) == "clear")
    }

    # a mut variable cannot be captured by the catch closure below
    let clear = $cleared

    if not $env._OMP_ENABLE_STREAMING {
        return (_omp_get_prompt primary $"--cleared=($clear)")
    }

    if $env._OMP_SERVE_JOB not-in (job list | get id) {
        _omp_start_serve
    }

    let context = (_omp_context)
    try { job kill $env._OMP_CYCLE }
    $env._OMP_CYCLE = (job spawn --auto-kill { generate { |_| commandline set-prompt (job recv); {next: true} } true | ignore })
    (
        $context
        | merge {command: render, id: $env._OMP_CYCLE, pwd: $env.PWD, cleared: $clear, env: (_omp_env)}
        | job send $env._OMP_SERVE_JOB
    )

    try {
        job recv --tag $env._OMP_CYCLE --timeout 5sec
    } catch {
        # the caught timeout overwrote the exit code of the user's command
        $env.LAST_EXIT_CODE = $context.status
        # a daemon that cannot answer would otherwise cost the timeout on every prompt
        $env._OMP_ENABLE_STREAMING = false
        try { job kill $env._OMP_SERVE_JOB }
        _omp_get_prompt primary $"--cleared=($clear)"
    }
}

def _omp_transient [] {
    if $env._OMP_ENABLE_STREAMING {
        let transient = (try { job recv --tag $env._OMP_CYCLE --timeout 100ms } catch { '' })
        if ($transient | is-not-empty) {
            return $transient
        }
    }
    _omp_get_prompt transient
}

def --env _omp_enable_streaming [] {
    $env._OMP_ENABLE_STREAMING = true
    $env._OMP_EXECUTED = false
    $env._OMP_CLEARED = $nu.history-enabled and (history | is-empty)
    $env.config.hooks.pre_execution = ($env.config.hooks.pre_execution | append {||
        let line = (commandline | str trim)
        # an empty Enter fires the hook too, but runs nothing
        if ($line | is-empty) {
            return
        }
        $env._OMP_EXECUTED = true
        $env._OMP_CLEARED = $nu.history-enabled and $line == 'clear'
    })
    $env.config.hooks.pre_prompt = ($env.config.hooks.pre_prompt | append {||
        $env._OMP_PRIMARY = (_omp_primary)
    })
    $env.PROMPT_COMMAND = {|| $env._OMP_PRIMARY? | default '' }
}

$env.PROMPT_COMMAND = {|| _omp_primary }
$env.PROMPT_COMMAND_RIGHT = {|| _omp_get_prompt right }

