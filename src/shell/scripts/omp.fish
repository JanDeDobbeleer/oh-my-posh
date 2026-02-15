set --export --global POSH_SHELL fish
set --export --global POSH_SHELL_VERSION $FISH_VERSION
set --export --global POWERLINE_COMMAND oh-my-posh
set --export --global CONDA_PROMPT_MODIFIER false

set --global _omp_tooltip_command ''
set --global _omp_current_rprompt ''
set --global _omp_transient 0
set --global _omp_executable ::OMP::
set --global _omp_ftcs_marks 0
set --global _omp_transient_prompt 0
set --global _omp_prompt_mark 0

# streaming support variables
set --global _omp_enable_streaming 0
set --global _omp_streaming_pid 0
set --global _omp_streaming_tempfile ''
set --global _omp_primary_prompt ''

# disable all known python virtual environment prompts
set --global VIRTUAL_ENV_DISABLE_PROMPT 1
set --global PYENV_VIRTUALENV_DISABLE_PROMPT 1

# We use this to avoid unnecessary CLI calls for prompt repaint.
set --global _omp_new_prompt 1

# template function for context loading
function set_poshcontext
    return
end

# cleanup stream resources
function _omp_cleanup_stream
    # kill background process if running
    if test -n "$_omp_streaming_pid" -a "$_omp_streaming_pid" -gt 0 2>/dev/null
        kill $_omp_streaming_pid 2>/dev/null
        set --global _omp_streaming_pid 0
    end
    # remove temp file
    if test -n "$_omp_streaming_tempfile"
        rm -f "$_omp_streaming_tempfile" 2>/dev/null
    end
end

# shell exit handler
function _omp_exit_handler --on-event fish_exit
    _omp_cleanup_stream
end

# streaming background reader: reads null-delimited prompts and signals parent
function _omp_streaming_reader
    set --local parent_pid $argv[1]
    set --local tempfile $argv[2]
    set --local count 0

    # read null-delimited prompts from oh-my-posh stream
    while read --null --local prompt
        # write to temp file (atomic via printf)
        printf '%s' "$prompt" >$tempfile

        # signal parent for updates after first prompt (index > 0)
        if test $count -gt 0
            kill -SIGUSR1 $parent_pid 2>/dev/null
        end

        set count (math $count + 1)
    end
end

# signal handler: called when streaming process has new prompt
function _omp_streaming_handler --on-signal SIGUSR1
    # only process if streaming is active
    if test $_omp_enable_streaming -eq 0
        return
    end

    if test -z "$_omp_streaming_tempfile" -o ! -f "$_omp_streaming_tempfile"
        return
    end

    # read updated prompt from temp file
    set --global _omp_primary_prompt (cat "$_omp_streaming_tempfile")
    set --global _omp_current_prompt "$_omp_primary_prompt"

    # trigger repaint
    commandline --function repaint
end

# start oh-my-posh stream process, block until first prompt arrives
function _omp_start_streaming
    # cleanup any prior stream
    _omp_cleanup_stream

    # determine temp file location
    set --local tmpdir $TMPDIR
    if test -z "$tmpdir"
        set tmpdir /tmp
    end
    set --global _omp_streaming_tempfile "$tmpdir/omp-fish-$fish_pid.txt"
    rm -f "$_omp_streaming_tempfile" 2>/dev/null

    # build stream command with all context
    set --local stream_cmd $_omp_executable stream \
        --shell=fish \
        --shell-version=$FISH_VERSION \
        --status=$_omp_status \
        --pipestatus="$_omp_pipestatus" \
        --no-status=$_omp_no_status \
        --execution-time=$_omp_execution_time \
        --stack-count=$_omp_stack_count

    # start background reader process
    $stream_cmd | _omp_streaming_reader $fish_pid "$_omp_streaming_tempfile" &
    set --global _omp_streaming_pid (jobs --last --pid)

    # block until first prompt arrives (mirrors zsh's blocking read)
    set --local timeout 5000
    set --local elapsed 0
    while not test -s "$_omp_streaming_tempfile"
        sleep 0.01
        set elapsed (math $elapsed + 10)
        if test $elapsed -ge $timeout
            # timeout - cleanup and return failure
            _omp_cleanup_stream
            return 1
        end
    end

    # read first prompt
    set --global _omp_primary_prompt (cat "$_omp_streaming_tempfile")
    set --global _omp_current_prompt "$_omp_primary_prompt"

    return 0
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

    # === STREAMING PATH ===
    if test $_omp_enable_streaming -eq 1
        # start new streaming session (blocks until first prompt arrives)
        if _omp_start_streaming
            # _omp_current_prompt already set by _omp_start_streaming
            echo -n "$_omp_current_prompt"
            return
        end
        # fall through to sync on failure
    end

    # === FALLBACK PATH ===
    # The prompt is saved for possible reuse, typically a repaint after clearing the screen buffer.
    set --global _omp_current_prompt (_omp_get_prompt primary --cleared=$omp_cleared | string join \n | string collect)

    echo -n "$_omp_current_prompt"
end

function fish_right_prompt
    if test "$_omp_transient" = 1
        set --global _omp_transient 0
        return
    end

    # Repaint an existing right prompt.
    if test "$_omp_new_prompt" = 0
        echo -n "$_omp_current_rprompt"
        return
    end

    set --global _omp_new_prompt 0
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
        set --global _omp_new_prompt 1
        set --global _omp_tooltip_command ''

        # cleanup streaming before executing command
        _omp_cleanup_stream

        if test $_omp_transient_prompt = 1
            set --global _omp_transient 1
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
    set --global _omp_new_prompt 1
    set --global _omp_tooltip_command ''

    # cleanup streaming before canceling
    _omp_cleanup_stream

    if test $_omp_transient_prompt = 1
        set --global _omp_transient 1
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
    set --global _omp_new_prompt 1
    commandline --function repaint
end
