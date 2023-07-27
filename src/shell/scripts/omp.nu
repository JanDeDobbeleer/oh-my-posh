export-env {
    $env.POWERLINE_COMMAND = 'oh-my-posh'
    $env.POSH_THEME = ::CONFIG::
    $env.PROMPT_INDICATOR = ""
    $env.POSH_PID = (random uuid)
    # By default displays the right prompt on the first line
    # making it annoying when you have a multiline prompt
    # making the behavior different compared to other shells
    $env.PROMPT_COMMAND_RIGHT = ''
    $env.POSH_SHELL_VERSION = (version | get version)

    # PROMPTS
    $env.PROMPT_MULTILINE_INDICATOR = (^::OMP:: print secondary $"--config=($env.POSH_THEME)" --shell=nu $"--shell-version=($env.POSH_SHELL_VERSION)")

    $env.PROMPT_COMMAND = { ||
        # We have to do this because the initial value of `$env.CMD_DURATION_MS` is always `0823`,
        # which is an official setting.
        # See https://github.com/nushell/nushell/discussions/6402#discussioncomment-3466687.
        let cmd_duration = if $env.CMD_DURATION_MS == "0823" { 0 } else { $env.CMD_DURATION_MS }

        # hack to set the cursor line to 1 when the user clears the screen
        # this obviously isn't bulletproof, but it's a start
        let clear = (history | last 1 | get 0.command) == "clear"

        let width = ((term size).columns | into string)
        ^::OMP:: print primary $"--config=($env.POSH_THEME)" --shell=nu $"--shell-version=($env.POSH_SHELL_VERSION)" $"--execution-time=($cmd_duration)" $"--status=($env.LAST_EXIT_CODE)" $"--terminal-width=($width)" $"--cleared=($clear)"
    }
}

if "::UPGRADE::" == "true" {
    echo "::UPGRADENOTICE::"
}
