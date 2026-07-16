set --export --global POSH_SHELL fish
set --export --global POSH_SHELL_VERSION $FISH_VERSION
set --export --global POWERLINE_COMMAND oh-my-posh
set --export --global CONDA_PROMPT_MODIFIER false

set --global _omp_tooltip_command ''
set --global _omp_current_rprompt ''
set --global _omp_transient 0
set --global _omp_executable ::OMP::
set --global _omp_cursor_positioning 0
set --global _omp_ftcs_marks 0
set --global _omp_transient_prompt 0
set --global _omp_transient_rprompt 0
set --global _omp_prompt_mark 0

# streaming support variables
set --global _omp_enable_streaming 0
set --global _omp_streaming_pid 0
set --global _omp_streaming_tempfile ''
set --global _omp_primary_prompt ''

# serve daemon variables
# A persistent `oh-my-posh serve` process renders prompts on request,
# replacing a process spawn per prompt with an in-memory render. Requests go
# through a named pipe (the only way fish can write to a running process);
# records flow back through a background reader into tempfiles.
set --global _omp_serve_pid 0

# _omp_serve_pid holds every pid of the daemon pipeline (fish's
# `jobs --last --pid` lists one per stage). The first is the daemon itself
# and serves as the liveness proxy; kill/stop target all of them.
#
# NEVER pass a possibly-zero pid to kill: `kill -0 0` signals the caller's
# own process group and always succeeds.
function _omp_serve_alive
    test -n "$_omp_serve_pid[1]" -a "$_omp_serve_pid[1]" -gt 0 2>/dev/null
    and kill -0 $_omp_serve_pid[1] 2>/dev/null
end
set --global _omp_serve_fifo ''
set --global _omp_serve_tempfile ''
set --global _omp_serve_cycle 0
set --global _omp_serve_failures 0

# disable all known python virtual environment prompts
set --global VIRTUAL_ENV_DISABLE_PROMPT 1
set --global PYENV_VIRTUALENV_DISABLE_PROMPT 1

# We use this to avoid unnecessary CLI calls for prompt repaint.
set --global _omp_new_prompt 1

function _omp_set_cursor_position
    # not supported in Midnight Commander
    # see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3415
    if test "$_omp_cursor_positioning" = 0; or set --query MC_SID
        return
    end

    set --local oldstty (stty -g </dev/tty)
    stty raw -echo min 1 </dev/tty

    set --local pos ''
    printf '\e[6n' >/dev/tty
    while true
        read --null --nchars 1 --local ch </dev/tty
        set pos $pos$ch
        string match -q 'R' $ch; and break
    end

    stty $oldstty </dev/tty

    set --local parts (string match -gr '\[(\d+);(\d+)R' $pos)
    set --export --global POSH_CURSOR_LINE $parts[1]
    set --export --global POSH_CURSOR_COLUMN $parts[2]
end

# template function for context loading
function set_poshcontext
    return
end

# cleanup stream resources
function _omp_cleanup_stream
    # kill the background pipeline if running; _omp_streaming_pid holds one
    # pid per pipeline stage, the first serves as the liveness proxy
    if test -n "$_omp_streaming_pid[1]" -a "$_omp_streaming_pid[1]" -gt 0 2>/dev/null
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
    _omp_serve_quit
    _omp_cleanup_stream
    # the transient prompt cache outlives stream cleanup by design, remove it on exit
    if test -n "$_omp_streaming_tempfile"
        rm -f "$_omp_streaming_tempfile.transient" 2>/dev/null
    end
end

# streaming background reader: reads null-delimited prompts and signals parent.
#
# IMPORTANT: this must run as an external `fish -c` process. fish does NOT
# fork a backgrounded pipeline stage that is a function - it runs in the main
# shell and blocks it. An external process pipeline forks properly, hence
# tempfiles + SIGUSR1 instead of variables.
set --global _omp_streaming_reader_script '
set parent_pid $argv[1]
set tempfile $argv[2]
set count 0
set marker (printf "\x1e")
while read --null prompt
    # a record prefixed with U+001E carries the transient prompt: cache it
    # for the key handlers so rendering the transient prompt needs no CLI call
    if test (string sub --length 1 -- $prompt | string collect) = "$marker"
        printf "%s" (string sub --start 2 -- $prompt | string collect) >"$tempfile.transient"
        continue
    end
    # write to temp file (atomic via printf)
    printf "%s" "$prompt" >$tempfile
    # signal parent for updates after first prompt (index > 0)
    if test $count -gt 0
        kill -SIGUSR1 $parent_pid 2>/dev/null
    end
    set count (math $count + 1)
end
'

# signal handler: called when streaming process has new prompt
function _omp_streaming_handler --on-signal SIGUSR1
    # only process if streaming is active
    if test $_omp_enable_streaming -eq 0
        return
    end

    # serve path: records carry a cycle id prefix; discard stale ones and
    # skip repaints for unchanged content (the synchronously consumed first
    # record also signals)
    if test -n "$_omp_serve_tempfile" -a -f "$_omp_serve_tempfile"; and _omp_serve_alive
        set --local us (printf '\x1f')
        set --local record (cat "$_omp_serve_tempfile" | string collect)
        set --local id (string split --max 1 -- $us $record)[1]
        if test "$id" != "$_omp_serve_cycle"
            return
        end

        set --local payload (string replace -- "$id$us" '' $record | string collect)
        if test "$payload" = "$_omp_primary_prompt"
            return
        end

        set --global _omp_primary_prompt $payload
        set --global _omp_current_prompt "$payload"
        commandline --function repaint
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
    # also invalidate the previous cycle's transient prompt
    rm -f "$_omp_streaming_tempfile" "$_omp_streaming_tempfile.transient" 2>/dev/null

    # build stream command with all context
    set --local stream_cmd $_omp_executable stream \
        --save-cache \
        --shell=fish \
        --shell-version=$FISH_VERSION \
        --status=$_omp_status \
        --pipestatus="$_omp_pipestatus" \
        --no-status=$_omp_no_status \
        --execution-time=$_omp_execution_time \
        --stack-count=$_omp_stack_count

    # start background reader process
    $stream_cmd | fish --no-config -c $_omp_streaming_reader_script $fish_pid "$_omp_streaming_tempfile" &
    # `jobs --last --pid` prints a header line plus one pid per pipeline
    # stage (fish 4.x) - keep only the numbers
    set --global _omp_streaming_pid (jobs --last --pid | string match --regex '^\d+$')

    # keep the stream pipeline out of job notifications
    disown $_omp_streaming_pid 2>/dev/null

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

# === serve daemon ===

# background reader for the daemon's records: "<id>\x1f<payload>\0", where a
# payload prefixed with \x1e carries the transient prompt.
#
# IMPORTANT: this must run as an external `fish -c` process. fish does NOT
# fork a backgrounded pipeline stage that is a function - it runs in the main
# shell and blocks it. An external process pipeline forks properly, hence
# tempfiles + SIGUSR1 instead of variables.
set --global _omp_serve_reader_script '
set parent_pid $argv[1]
set tempfile $argv[2]
set marker (printf "\x1e")
while read --null record
    set m (string match --regex -- "^(\d+)\x1f" $record)
    if test (count $m) -ne 2
        continue
    end
    set prefix_len (string length -- "$m[1]")
    set payload (string sub --start (math $prefix_len + 1) -- $record | string collect)
    # transient record: cache for the key handlers, never repaint
    if test (string sub --length 1 -- $payload | string collect) = "$marker"
        printf "%s" (string sub --start 2 -- $payload | string collect) >"$tempfile.transient"
        continue
    end
    # keep the id with the payload so consumers can discard stale records
    printf "%s\x1f%s" $m[2] $payload >$tempfile
    # wake the parent; its SIGUSR1 handler dedupes unchanged content, so
    # signaling the synchronously consumed first record is harmless
    kill -SIGUSR1 $parent_pid 2>/dev/null
end
'

function _omp_serve_stop
    if test -n "$_omp_serve_pid[1]" -a "$_omp_serve_pid[1]" -gt 0 2>/dev/null
        kill $_omp_serve_pid 2>/dev/null
    end
    set --global _omp_serve_pid 0

    if test -n "$_omp_serve_fifo"
        rm -f "$_omp_serve_fifo" 2>/dev/null
        set --global _omp_serve_fifo ''
    end
    # the tempfiles survive on purpose: the .transient cache must outlive the
    # cycle so the queued transient repaint can still use it
end

function _omp_serve_start
    _omp_serve_stop

    set --local tmpdir $TMPDIR
    if test -z "$tmpdir"
        set tmpdir /tmp
    end

    set --global _omp_serve_fifo "$tmpdir/omp-serve-$fish_pid.req"
    set --global _omp_serve_tempfile "$tmpdir/omp-serve-$fish_pid.txt"
    rm -f "$_omp_serve_fifo" "$_omp_serve_tempfile" "$_omp_serve_tempfile.transient" 2>/dev/null

    if not mkfifo -m 600 "$_omp_serve_fifo" 2>/dev/null
        return 1
    end

    # The daemon opens the fifo itself read-write, so our open-write-close
    # per request never EOFs it. Its stderr must never reach the terminal
    # (a Go panic would corrupt the display), and it self-terminates when its
    # stdout reader disappears (broken pipe on the next record write).
    $_omp_executable serve --shell=fish --request-pipe "$_omp_serve_fifo" 2>/dev/null | fish --no-config -c $_omp_serve_reader_script $fish_pid "$_omp_serve_tempfile" &
    # `jobs --last --pid` prints a header line plus one pid per pipeline
    # stage (fish 4.x) - keep only the numbers
    set --global _omp_serve_pid (jobs --last --pid | string match --regex '^\d+$')

    if not _omp_serve_alive
        _omp_serve_stop
        return 1
    end

    # keep the daemon pipeline out of job notifications
    disown $_omp_serve_pid 2>/dev/null

    return 0
end

# JSON string escaping with builtins. Joining the result with '' drops any
# embedded newlines - the values we send (paths, POSH_* variables) never
# legitimately contain them, and a raw newline would break the
# line-delimited request protocol.
function _omp_serve_escape
    string join '' -- (string replace -a -- '\\' '\\\\' "$argv" | string replace -a -- '"' '\\"' | string replace -a -- \t '\\t')
end

function _omp_serve_env_json
    set --local parts

    set --append parts '"PATH":"'(_omp_serve_escape (string join ':' -- $PATH))'"'

    # forward every exported POSH_* variable plus the virtual-env markers;
    # the daemon's environment is otherwise frozen at its start
    for name in (set --names)
        if string match -q 'POSH_*' -- $name; and set -qx $name
            set --append parts '"'$name'":"'(_omp_serve_escape "$$name")'"'
        end
    end

    for name in VIRTUAL_ENV CONDA_PROMPT_MODIFIER
        if set -q $name
            set --append parts '"'$name'":"'(_omp_serve_escape "$$name")'"'
        end
    end

    echo -n '{'(string join ',' -- $parts)'}'
end

function _omp_serve_request
    set --global _omp_serve_cycle (math $_omp_serve_cycle + 1)
    rm -f "$_omp_serve_tempfile" "$_omp_serve_tempfile.transient" 2>/dev/null

    set --local exec_time $_omp_execution_time
    if test -z "$exec_time"
        set exec_time 0
    end

    set --local width $COLUMNS
    if test -z "$width"
        set width 0
    end

    set --local cleared $_omp_cleared
    if test -z "$cleared"
        set cleared false
    end

    set --local json '{"command":"render","id":'$_omp_serve_cycle',"shell":"fish","shell-version":"'$FISH_VERSION'","status":'$_omp_status',"pipestatus":"'"$_omp_pipestatus"'","no-status":'$_omp_no_status',"execution-time":'$exec_time',"stack-count":'(count $dirstack)',"terminal-width":'$width',"cleared":'$cleared',"pwd":"'(_omp_serve_escape $PWD)'","env":'(_omp_serve_env_json)'}'

    # a fifo write with no reader blocks forever - only write while the
    # daemon pipeline is alive (the daemon holds the fifo open read-write,
    # so a live daemon can never block us)
    if not _omp_serve_alive
        return 1
    end

    echo $json >"$_omp_serve_fifo" 2>/dev/null
end

# poll the tempfile until it holds this cycle's primary record
function _omp_serve_read_primary
    set --local us (printf '\x1f')
    set --local timeout 2000
    set --local elapsed 0

    while test $elapsed -lt $timeout
        if test -s "$_omp_serve_tempfile"
            set --local record (cat "$_omp_serve_tempfile" | string collect)
            set --local id (string split --max 1 -- $us $record)[1]
            if test "$id" = "$_omp_serve_cycle"
                string replace -- "$id$us" '' $record | string collect
                return 0
            end
        end
        sleep 0.01
        set elapsed (math $elapsed + 10)
    end

    return 1
end

# renders the primary prompt through the daemon; returns nonzero on failure,
# in which case the caller falls back to the per-prompt stream. The transient
# prompt lands in "$_omp_serve_tempfile.transient" via the reader.
function _omp_serve_render
    if not _omp_serve_alive
        if not _omp_serve_start
            set --global _omp_serve_failures (math $_omp_serve_failures + 1)
            return 1
        end
    end

    if not _omp_serve_request
        # the daemon died since the last prompt - restart it once
        if not _omp_serve_start; or not _omp_serve_request
            set --global _omp_serve_failures (math $_omp_serve_failures + 1)
            _omp_serve_stop
            return 1
        end
    end

    set --local prompt (_omp_serve_read_primary)
    if test -z "$prompt"
        set --global _omp_serve_failures (math $_omp_serve_failures + 1)
        _omp_serve_stop
        return 1
    end

    set --global _omp_primary_prompt $prompt
    set --global _omp_current_prompt "$prompt"
    return 0
end

function _omp_serve_abort
    if test -n "$_omp_serve_fifo"; and _omp_serve_alive
        echo '{"command":"abort"}' >"$_omp_serve_fifo" 2>/dev/null
    end
end

function _omp_serve_quit
    if test -n "$_omp_serve_fifo"; and _omp_serve_alive
        echo '{"command":"quit"}' >"$_omp_serve_fifo" 2>/dev/null
    end
    _omp_serve_stop
    if test -n "$_omp_serve_tempfile"
        rm -f "$_omp_serve_tempfile" "$_omp_serve_tempfile.transient" 2>/dev/null
        set --global _omp_serve_tempfile ''
    end
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
        # prefer the transient prompt rendered ahead of time by the serve
        # daemon or the streaming process, saves a CLI call
        if test -n "$_omp_serve_tempfile"; and test -s "$_omp_serve_tempfile.transient"
            cat "$_omp_serve_tempfile.transient"
        else if test $_omp_enable_streaming -eq 1; and test -n "$_omp_streaming_tempfile"; and test -s "$_omp_streaming_tempfile.transient"
            cat "$_omp_streaming_tempfile.transient"
        else
            _omp_get_prompt transient
        end
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
    _omp_set_cursor_position

    # validate if the user cleared the screen; global so the serve daemon
    # request builder sees it too
    set --global _omp_cleared false
    set --local last_command (history search --max 1)

    if test "$last_command" = clear
        set --global _omp_cleared true
    end

    if test $_omp_prompt_mark = 1
        iterm2_prompt_mark
    end

    # === STREAMING PATH ===
    if test $_omp_enable_streaming -eq 1
        # serve daemon: persistent process, no spawn per prompt. After three
        # failures the daemon is left alone for the session and every prompt
        # takes the per-prompt stream below instead.
        if test $_omp_serve_failures -lt 3; and _omp_serve_render
            echo -n "$_omp_current_prompt"
            return
        end

        # per-prompt stream (also the serve fallback; blocks until first prompt arrives)
        if _omp_start_streaming
            # _omp_current_prompt already set by _omp_start_streaming
            echo -n "$_omp_current_prompt"
            return
        end
        # fall through to sync on failure
    end

    # === FALLBACK PATH ===
    # The prompt is saved for possible reuse, typically a repaint after clearing the screen buffer.
    set --global _omp_current_prompt (_omp_get_prompt primary --cleared=$_omp_cleared | string join \n | string collect)

    echo -n "$_omp_current_prompt"
end

function fish_right_prompt
    if test "$_omp_transient" = 1
        set --global _omp_transient 0

        if test $_omp_transient_rprompt = 1
            _omp_get_prompt transient-right
        end

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
    if test $_omp_ftcs_marks != 1
        return
    end

    if test -n "$argv"
        # advertise the command line via kitty's cmdline_url= extension
        echo -ne "\e]133;C;cmdline_url="(string escape --style=url -- "$argv")"\a"
        return
    end

    echo -ne "\e]133;C\a"
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

function _omp_backspace_key_handler
    commandline --function backward-delete-char

    if test -z "$_omp_tooltip_command"
        return
    end

    set --local current_command (commandline --current-buffer | string trim -l | string split --allow-empty -f1 ' ' | string collect)

    if test "$current_command" = "$_omp_tooltip_command"
        return
    end

    set _omp_tooltip_command "$current_command"
    set _omp_current_rprompt (_omp_get_prompt tooltip --command="$current_command" | string join '')
    commandline --function repaint
end

function enable_poshtooltips
    bind \x20 _omp_space_key_handler -M default
    bind \x20 _omp_space_key_handler -M insert
    bind \x7f _omp_backspace_key_handler -M default
    bind \x7f _omp_backspace_key_handler -M insert
end

# transient prompt

function _omp_enter_key_handler
    if commandline --paging-mode
        set --global _omp_new_prompt 1
        set --global _omp_tooltip_command ''

        # stop prompt updates before executing: kill the per-prompt stream,
        # tell the daemon to abort its in-flight cycle (the daemon lives on)
        _omp_cleanup_stream
        _omp_serve_abort

        if test $_omp_transient_prompt = 1
            set --global _omp_transient 1
            commandline --function repaint
        end

        commandline --function execute
        return
    end

    if commandline --is-valid || test -z (commandline --current-buffer | string trim -l | string collect)
        set --global _omp_new_prompt 1
        set --global _omp_tooltip_command ''

        # stop prompt updates before executing: kill the per-prompt stream,
        # tell the daemon to abort its in-flight cycle (the daemon lives on)
        _omp_cleanup_stream
        _omp_serve_abort

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

    # stop prompt updates before canceling (the daemon lives on)
    _omp_cleanup_stream
    _omp_serve_abort

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
