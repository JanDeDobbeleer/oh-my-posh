set --export POSH_THEME ::CONFIG::
set --export POSH_SHELL_VERSION $FISH_VERSION
set --global POWERLINE_COMMAND oh-my-posh
set --global POSH_PID $fish_pid
set --global CONDA_PROMPT_MODIFIER false
set --global omp_tooltip_command ''
set --global omp_current_rprompt ''
set --global omp_transient false

# We use this to avoid unnecessary CLI calls for prompt repaint.
set --global omp_new_prompt true

# template function for context loading
function set_poshcontext
    return
end

# NOTE: Input function calls via `commandline --function` are put into a queue and will not be executed until an outer regular function returns. See https://fishshell.com/docs/current/cmds/commandline.html.

function fish_prompt
    set --local omp_status_cache_temp $status
    set --local omp_pipestatus_cache_temp $pipestatus
    # clear from cursor to end of screen as
    # commandline --function repaint does not do this
    # see https://github.com/fish-shell/fish-shell/issues/8418
    printf \e\[0J
    if test "$omp_transient" = true
        ::OMP:: print transient --config $POSH_THEME --shell fish --status $omp_status_cache --pipestatus="$omp_pipestatus_cache" --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION --no-status=$omp_no_exit_code
        return
    end
    if test "$omp_new_prompt" = false
        echo -n "$omp_current_prompt"
        return
    end
    set --global omp_status_cache $omp_status_cache_temp
    set --global omp_pipestatus_cache $omp_pipestatus_cache_temp
    set --global omp_stack_count (count $dirstack)
    set --global omp_duration "$CMD_DURATION$cmd_duration"
    set --global omp_no_exit_code false

    # check if variable set, < 3.2 case
    if set --query omp_lastcommand; and test -z "$omp_lastcommand"
        set omp_duration 0
        set omp_no_exit_code true
    end

    # works with fish >=3.2
    if set --query omp_last_status_generation; and test "$omp_last_status_generation" = "$status_generation"
        set omp_duration 0
        set omp_no_exit_code true
    else if test -z "$omp_last_status_generation"
        # first execution - $status_generation is 0, $omp_last_status_generation is empty
        set omp_no_exit_code true
    end

    if set --query status_generation
        set --global omp_last_status_generation $status_generation
    end

    set_poshcontext

    # validate if the user cleared the screen
    set --local omp_cleared false
    set --local last_command (history search --max 1)

    if test "$last_command" = clear
        set omp_cleared true
    end

    ::PROMPT_MARK::

    # The prompt is saved for possible reuse, typically a repaint after clearing the screen buffer.
    set --global omp_current_prompt (::OMP:: print primary --config $POSH_THEME --shell fish --status $omp_status_cache --pipestatus="$omp_pipestatus_cache" --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION --cleared=$omp_cleared --no-status=$omp_no_exit_code | string collect)
    echo -n "$omp_current_prompt"
end

function fish_right_prompt
    if test "$omp_transient" = true
        set omp_transient false
        return
    end
    # Repaint an existing right prompt.
    if test "$omp_new_prompt" = false
        echo -n "$omp_current_rprompt"
        return
    end
    set omp_new_prompt false
    set --global omp_current_rprompt (::OMP:: print right --config $POSH_THEME --shell fish --status $omp_status_cache --pipestatus="$omp_pipestatus_cache" --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION --no-status=$omp_no_exit_code | string join '')
    echo -n "$omp_current_rprompt"
end

function postexec_omp --on-event fish_postexec
    # works with fish <3.2
    # pre and postexec not fired for empty command in fish >=3.2
    set --global omp_lastcommand $argv
end

function preexec_omp --on-event fish_preexec
    if test "::FTCS_MARKS::" = true
        echo -ne "\e]133;C\a"
    end
end

# perform cleanup so a new initialization in current session works
if test -n (bind \r --user 2>/dev/null | string match -e _omp_enter_key_handler)
    bind -e \r -M default
    bind -e \r -M insert
    bind -e \r -M visual
end
if test -n (bind \n --user 2>/dev/null | string match -e _omp_enter_key_handler)
    bind -e \n -M default
    bind -e \n -M insert
    bind -e \n -M visual
end
if test -n (bind \cc --user 2>/dev/null | string match -e _omp_ctrl_c_key_handler)
    bind -e \cc -M default
    bind -e \cc -M insert
    bind -e \cc -M visual
end
if test -n (bind \x20 --user 2>/dev/null | string match -e _omp_space_key_handler)
    bind -e \x20 -M default
    bind -e \x20 -M insert
end

# tooltip

function _omp_space_key_handler
    commandline --function expand-abbr
    commandline --insert ' '
    # Get the first word of command line as tip.
    set --local tooltip_command (commandline --current-buffer | string trim -l | string split --allow-empty -f1 ' ' | string collect)
    # Ignore an empty/repeated tooltip command.
    if test -z "$tooltip_command" || test "$tooltip_command" = "$omp_tooltip_command"
        return
    end
    set omp_tooltip_command $tooltip_command
    set --local tooltip_prompt (::OMP:: print tooltip --config $POSH_THEME --shell fish --status $omp_status_cache --pipestatus="$omp_pipestatus_cache" --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION --command $omp_tooltip_command --no-status=$omp_no_exit_code | string join '')
    if test -z "$tooltip_prompt"
        return
    end
    # Save the tooltip prompt to avoid unnecessary CLI calls.
    set omp_current_rprompt $tooltip_prompt
    commandline --function repaint
end

if test "::TOOLTIPS::" = true
    bind \x20 _omp_space_key_handler -M default
    bind \x20 _omp_space_key_handler -M insert
end

# transient prompt

function _omp_enter_key_handler
    if commandline --paging-mode
        commandline --function accept-autosuggestion
        return
    end
    if commandline --is-valid || test -z (commandline --current-buffer | string trim -l | string collect)
        set omp_new_prompt true
        set omp_tooltip_command ''
        if test "::TRANSIENT::" = true
            set omp_transient true
            commandline --function repaint
        end
    end
    commandline --function execute
end

function _omp_ctrl_c_key_handler
    if test -z (commandline --current-buffer | string collect)
        return
    end
    # Render a transient prompt on Ctrl-C with non-empty command line buffer.
    set omp_new_prompt true
    set omp_tooltip_command ''
    if test "::TRANSIENT::" = true
        set omp_transient true
        commandline --function repaint
    end
    commandline --function cancel-commandline
    commandline --function repaint
end

bind \r _omp_enter_key_handler -M default
bind \r _omp_enter_key_handler -M insert
bind \r _omp_enter_key_handler -M visual
bind \n _omp_enter_key_handler -M default
bind \n _omp_enter_key_handler -M insert
bind \n _omp_enter_key_handler -M visual
bind \cc _omp_ctrl_c_key_handler -M default
bind \cc _omp_ctrl_c_key_handler -M insert
bind \cc _omp_ctrl_c_key_handler -M visual

# legacy functions
function enable_poshtooltips
    return
end
function enable_poshtransientprompt
    return
end

if test "::UPGRADE::" = true
    echo "::UPGRADENOTICE::"
end

if test "::AUTOUPGRADE::" = true
    ::OMP:: upgrade
end
