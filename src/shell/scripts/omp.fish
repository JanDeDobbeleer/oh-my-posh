set --export POSH_THEME ::CONFIG::
set --export POSH_SHELL fish
set --export POSH_SHELL_VERSION $FISH_VERSION
set --export POWERLINE_COMMAND oh-my-posh
set --export POSH_SESSION_ID ::SESSION_ID::
set --export CONDA_PROMPT_MODIFIER false

set --global _omp_tooltip_command ''
set --global _omp_current_rprompt ''
set --global _omp_transient 0
set --global _omp_executable ::OMP::
set --global _omp_ftcs_marks 0
set --global _omp_transient_prompt 0
set --global _omp_prompt_mark 0

# disable all known python virtual environment prompts
set --global VIRTUAL_ENV_DISABLE_PROMPT 1
set --global PYENV_VIRTUALENV_DISABLE_PROMPT 1

# We use this to avoid unnecessary CLI calls for prompt repaint.
set --global _omp_new_prompt 1

# template function for context loading
function set_poshcontext
    return
end

function _omp_get_prompt
    if test (count $argv) -eq 0
        return
    end
    $_omp_executable print $argv[1] \
        --save-cache \
        --shell=fish \
        --shell-version=$FISH_VERSION \
        --status=$_omp_status \
        --pipestatus="$_omp_pipestatus" \
        --no-status=$_omp_no_status \
        --execution-time=$_omp_execution_time \
        --stack-count=$_omp_stack_count \
        $argv[2..]
end

# NOTE: Input function calls via `commandline --function` are put into a queue and will not be executed until an outer regular function returns. See https://fishshell.com/docs/current/cmds/commandline.html.

function fish_prompt
    set --local omp_status_temp $status
    set --local omp_pipestatus_temp $pipestatus
    # clear from cursor to end of screen as
    # commandline --function repaint does not do this
    # see https://github.com/fish-shell/fish-shell/issues/8418
    printf \e\[0J
    if test "$_omp_transient" = 1
        _omp_get_prompt transient
        return
    end
    if test "$_omp_new_prompt" = 0
        echo -n "$_omp_current_prompt"
        return
    end
    set --global _omp_status $omp_status_temp
    set --global _omp_pipestatus $omp_pipestatus_temp
    set --global _omp_no_status false
    set --global _omp_execution_time "$CMD_DURATION$cmd_duration"
    set --global _omp_stack_count (count $dirstack)

    # check if variable set, < 3.2 case
    if set --query _omp_last_command && test -z "$_omp_last_command"
        set _omp_execution_time 0
        set _omp_no_status true
    end

    # works with fish >=3.2
    if set --query _omp_last_status_generation && test "$_omp_last_status_generation" = "$status_generation"
        set _omp_execution_time 0
        set _omp_no_status true
    else if test -z "$_omp_last_status_generation"
        # first execution - $status_generation is 0, $_omp_last_status_generation is empty
        set _omp_no_status true
    end

    if set --query status_generation
        set --global _omp_last_status_generation $status_generation
    end

    set_poshcontext

    # validate if the user cleared the screen
    set --local omp_cleared false
    set --local last_command (history search --max 1)

    if test "$last_command" = clear
        set omp_cleared true
    end

    if test $_omp_prompt_mark = 1
        iterm2_prompt_mark
    end

    # The prompt is saved for possible reuse, typically a repaint after clearing the screen buffer.
    set --global _omp_current_prompt (_omp_get_prompt primary --cleared=$omp_cleared | string join \n | string collect)

    echo -n "$_omp_current_prompt"
end

function fish_right_prompt
    if test "$_omp_transient" = 1
        set _omp_transient 0
        return
    end

    # Repaint an existing right prompt.
    if test "$_omp_new_prompt" = 0
        echo -n "$_omp_current_rprompt"
        return
    end

    set _omp_new_prompt 0
    set --global _omp_current_rprompt (_omp_get_prompt right | string join '')

    echo -n "$_omp_current_rprompt"
end

function _omp_postexec --on-event fish_postexec
    # works with fish <3.2
    # pre and postexec not fired for empty command in fish >=3.2
    set --global _omp_last_command $argv
end

function _omp_preexec --on-event fish_preexec
    if test $_omp_ftcs_marks = 1
        echo -ne "\e]133;C\a"
    end
end

# perform cleanup so a new initialization in current session works
if bind \r --user 2>/dev/null | string match -qe _omp_enter_key_handler
    bind -e \r -M default
    bind -e \r -M insert
    bind -e \r -M visual
end

if bind \n --user 2>/dev/null | string match -qe _omp_enter_key_handler
    bind -e \n -M default
    bind -e \n -M insert
    bind -e \n -M visual
end

if bind \cc --user 2>/dev/null | string match -qe _omp_ctrl_c_key_handler
    bind -e \cc -M default
    bind -e \cc -M insert
    bind -e \cc -M visual
end

if bind \x20 --user 2>/dev/null | string match -qe _omp_space_key_handler
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
    if test -z "$tooltip_command" || test "$tooltip_command" = "$_omp_tooltip_command"
        return
    end

    set _omp_tooltip_command $tooltip_command
    set --local tooltip_prompt (_omp_get_prompt tooltip --command=$_omp_tooltip_command | string join '')

    if test -z "$tooltip_prompt"
        return
    end

    # Save the tooltip prompt to avoid unnecessary CLI calls.
    set _omp_current_rprompt $tooltip_prompt
    commandline --function repaint
end

function enable_poshtooltips
    bind \x20 _omp_space_key_handler -M default
    bind \x20 _omp_space_key_handler -M insert
end

# transient prompt

function _omp_enter_key_handler
    if commandline --paging-mode
        commandline --function execute
        return
    end

    if commandline --is-valid || test -z (commandline --current-buffer | string trim -l | string collect)
        set _omp_new_prompt 1
        set _omp_tooltip_command ''

        if test $_omp_transient_prompt = 1
            set _omp_transient 1
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
    set _omp_new_prompt 1
    set _omp_tooltip_command ''

    if test $_omp_transient_prompt = 1
        set _omp_transient 1
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
function enable_poshtransientprompt
    return
end

# This can be called by user whenever re-rendering is required.
function omp_repaint_prompt
    set _omp_new_prompt 1
    commandline --function repaint
end
