export POSH_SHELL='zsh'
export POSH_SHELL_VERSION=$ZSH_VERSION
export POWERLINE_COMMAND='oh-my-posh'
export CONDA_PROMPT_MODIFIER=false
export ZLE_RPROMPT_INDENT=0
export OSTYPE=$OSTYPE

# disable all known python virtual environment prompts
export VIRTUAL_ENV_DISABLE_PROMPT=1
export PYENV_VIRTUALENV_DISABLE_PROMPT=1

_omp_executable=::OMP::
_omp_tooltip_command=''

# zsh/datetime provides the epochtime array for native millisecond timestamps
zmodload zsh/datetime 2>/dev/null

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# streaming support variables
# Preserve the current fd value if the script is re-sourced mid-session; default to -1 when unset or empty.
_omp_stream_fd=${_omp_stream_fd:--1}
_omp_enable_streaming=0
_omp_primary_prompt=""
_omp_transient_prompt=""
_omp_streaming_supported=""

# serve daemon variables
# A persistent `oh-my-posh serve` process renders prompts on request, replacing
# a process spawn per prompt with an in-memory render. Preserve the fds when
# the script is re-sourced mid-session so the running daemon is reused.
_omp_serve_fd_in=${_omp_serve_fd_in:--1}
_omp_serve_fd_out=${_omp_serve_fd_out:--1}
_omp_serve_pid=${_omp_serve_pid:-0}
_omp_serve_cycle=0
_omp_serve_failures=0

# set secondary prompt (also exports POSH_MULTILINE_KEEPPROMPT)
eval "$($_omp_executable print secondary --shell=zsh --eval)"

function _omp_set_cursor_position() {
  # not supported in Midnight Commander
  # see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3415
  if [[ $_omp_cursor_positioning == 0 ]] || [[ -v MC_SID ]]; then
    return
  fi

  local oldstty=$(stty -g)
  stty raw -echo min 0

  local pos
  echo -en '\033[6n' >/dev/tty
  read -r -d R pos
  pos=${pos:2} # strip off the esc-[
  local parts=(${(s:;:)pos})

  stty $oldstty

  export POSH_CURSOR_LINE=${parts[1]}
  export POSH_CURSOR_COLUMN=${parts[2]}
}

# template function for context loading
function set_poshcontext() {
  return
}

# cleanup stream resources
function _omp_cleanup_stream() {
  # unregister handler first (prevents handler firing on closed fd)
  [[ $_omp_stream_fd -ge 0 ]] && zle -F $_omp_stream_fd 2>/dev/null
  # close fd — process gets SIGPIPE and terminates
  [[ $_omp_stream_fd -ge 0 ]] && eval "exec {_omp_stream_fd}<&-" 2>/dev/null
  _omp_stream_fd=-1
}

# start oh-my-posh stream process, block until first prompt arrives
function _omp_start_streaming() {
  # cleanup any stale streams
  _omp_cleanup_stream

  # the transient prompt for this cycle streams in alongside the primary
  # prompt updates, invalidate the previous cycle's version
  _omp_transient_prompt=""

  # build command with all context
  local -a stream_cmd=(
    "$_omp_executable" stream
    --save-cache
    --shell=zsh
    --shell-version="$ZSH_VERSION"
    --status=$_omp_status
    --pipestatus="${_omp_pipestatus[*]}"
    --no-status=$_omp_no_status
    --execution-time=$_omp_execution_time
    --job-count=$_omp_job_count
    --stack-count=$_omp_stack_count
    --terminal-width="${COLUMNS:-0}"
  )

  # start process substitution — no PID tracking needed
  # closing fd sends SIGPIPE which terminates oh-my-posh
  # redirect stream process stdin to prevent it from interfering with command output
  exec {_omp_stream_fd}< <(exec "${stream_cmd[@]}" </dev/null 2>/dev/null)

  if [[ $_omp_stream_fd -lt 0 ]]; then
    return 1
  fi

  # block until first prompt arrives (mirrors PowerShell's polling loop)
  IFS= read -r -u $_omp_stream_fd -d $'\0' _omp_primary_prompt
  if [[ $? -ne 0 || -z "$_omp_primary_prompt" ]]; then
    _omp_cleanup_stream
    return 1
  fi

  PS1="$_omp_primary_prompt"

  return 0
}

# async handler: called when data available on fd (reads single value)
function _omp_async_handler() {
  local fd=$1
  local record

  # read single null-delimited record (stream emits one value per \0)
  IFS= read -r -u $fd -d $'\0' record
  if [[ $? -ne 0 ]]; then
    if [[ -z "$record" ]]; then
      # EOF — only clean up if this is still the active stream fd.
      # A stale handler from a previous prompt cycle must not close the current stream.
      [[ $_omp_stream_fd -eq $fd ]] && _omp_cleanup_stream
      return 0
    fi
  fi

  # a record prefixed with U+001E carries the transient prompt: cache it for
  # the line-init widget so rendering the transient prompt needs no CLI call
  if [[ $record == $'\x1e'* ]]; then
    _omp_transient_prompt=${record#$'\x1e'}
    return 0
  fi

  _omp_primary_prompt=$record
  PS1="$_omp_primary_prompt"
  zle reset-prompt 2>/dev/null

  return 0
}

# === serve daemon ===

function _omp_serve_stop() {
  # unregister the zle watcher before closing its fd
  [[ $_omp_serve_fd_out -ge 0 ]] && zle -F $_omp_serve_fd_out 2>/dev/null
  [[ $_omp_serve_fd_in -ge 0 ]] && eval "exec {_omp_serve_fd_in}>&-" 2>/dev/null
  [[ $_omp_serve_fd_out -ge 0 ]] && eval "exec {_omp_serve_fd_out}<&-" 2>/dev/null
  _omp_serve_fd_in=-1
  _omp_serve_fd_out=-1
}

function _omp_serve_start() {
  _omp_serve_stop

  # With job control active, an interactive zsh announces the coproc on the
  # terminal like any background job ("[1] 12345"). Disabling MONITOR for the
  # spawn suppresses the notice; as a side effect the daemon starts with
  # SIGINT/SIGQUIT ignored (POSIX no-job-control semantics), so Ctrl+C at the
  # prompt cannot take it down.
  setopt localoptions no_monitor

  # The daemon's stderr must never reach the terminal (a Go panic would
  # corrupt the display).
  coproc "$_omp_executable" serve --shell=zsh 2>/dev/null
  [[ $? -ne 0 ]] && return 1
  _omp_serve_pid=$!

  # Duplicate both directions to session fds: the duplicates survive a later
  # `coproc` (ours or the user's) replacing the coproc slot. The stderr
  # redirect (suppressing "no coprocess" when the daemon failed to start) must
  # be scoped to a block: on a redirection-only `exec`, zsh applies every
  # listed redirection to the shell permanently, which would silence the
  # session's stderr for good.
  { exec {_omp_serve_fd_out}<&p {_omp_serve_fd_in}>&p } 2>/dev/null
  if [[ $_omp_serve_fd_in -lt 0 || $_omp_serve_fd_out -lt 0 ]]; then
    _omp_serve_stop
    return 1
  fi

  # Keep the daemon out of the job table: it must not show up in `jobs` or
  # inflate the job-count segment. Lifetime is governed by the fds - closing
  # them (or the shell dying) EOFs the daemon's stdin and it exits.
  disown %+ 2>/dev/null

  # One session-long watcher delivers async records (segment updates and the
  # transient refresh) while zle is active. It never races the synchronous
  # read in _omp_serve_render: zle - and therefore this watcher - is not
  # active while precmd runs.
  zle -F $_omp_serve_fd_out _omp_serve_async_handler 2>/dev/null

  return 0
}

function _omp_serve_escape() {
  # JSON string escaping using native parameter expansion; the result is
  # returned in REPLY. Any control characters left after the named escapes
  # are stripped - JSON forbids them raw.
  local s=$1
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  s=${s//$'\r'/\\r}
  s=${s//$'\t'/\\t}
  REPLY=${s//[[:cntrl:]]/}
}

function _omp_serve_request() {
  # A write to a dead daemon's pipe raises SIGPIPE, which kills a
  # non-interactive shell outright - ignore it for the duration of this
  # function (localtraps restores the user's disposition on return) so the
  # write degrades into the error path instead. The pid pre-check makes the
  # common case cheap; the trap covers the race.
  setopt localoptions localtraps
  trap '' PIPE

  # never pass a possibly-zero pid to kill: `kill -0 0` signals the caller's
  # own process group and always succeeds
  [[ $_omp_serve_pid -gt 0 ]] || return 1
  kill -0 $_omp_serve_pid 2>/dev/null || return 1

  (( _omp_serve_cycle++ ))
  _omp_transient_prompt=""

  local name env_json
  _omp_serve_escape "$PATH"
  env_json="\"PATH\":\"$REPLY\""

  # Forward every exported POSH_* variable plus the virtual-env markers; the
  # daemon's environment is otherwise frozen at its start.
  for name in ${(k)parameters[(I)POSH_*]}; do
    [[ ${parameters[$name]} == *export* ]] || continue
    _omp_serve_escape "${(P)name}"
    env_json+=",\"$name\":\"$REPLY\""
  done

  for name in VIRTUAL_ENV CONDA_PROMPT_MODIFIER; do
    [[ -v $name ]] || continue
    _omp_serve_escape "${(P)name}"
    env_json+=",\"$name\":\"$REPLY\""
  done

  _omp_serve_escape "$PWD"

  local json='{"command":"render"'
  json+=",\"id\":$_omp_serve_cycle"
  json+=',"shell":"zsh"'
  json+=",\"shell-version\":\"$ZSH_VERSION\""
  json+=",\"status\":$_omp_status"
  json+=",\"pipestatus\":\"${_omp_pipestatus[*]}\""
  json+=",\"no-status\":$_omp_no_status"
  json+=",\"execution-time\":$_omp_execution_time"
  json+=",\"stack-count\":$_omp_stack_count"
  json+=",\"terminal-width\":${COLUMNS:-0}"
  json+=",\"job-count\":$_omp_job_count"
  json+=",\"pwd\":\"$REPLY\""
  json+=",\"env\":{$env_json}"
  json+='}'

  print -r -u $_omp_serve_fd_in -- "$json" 2>/dev/null
}

# Renders the primary prompt through the daemon. Returns nonzero on failure,
# in which case the caller falls back to the per-prompt stream. The transient
# prompt records are cached into $_omp_transient_prompt, either here or by the
# async watcher once zle is active.
function _omp_serve_render() {
  if [[ $_omp_serve_fd_in -lt 0 ]] && ! _omp_serve_start; then
    (( _omp_serve_failures++ ))
    return 1
  fi

  if ! _omp_serve_request; then
    # The daemon died since the last prompt - restart it once.
    if ! _omp_serve_start || ! _omp_serve_request; then
      (( _omp_serve_failures++ ))
      _omp_serve_stop
      return 1
    fi
  fi

  # Block until this cycle's first primary record; async updates and the
  # transient refresh arrive later through the zle watcher. Stale records
  # from an aborted cycle are discarded by the id check.
  local record id payload
  while true; do
    if ! IFS= read -r -u $_omp_serve_fd_out -d $'\0' -t 2 record; then
      (( _omp_serve_failures++ ))
      _omp_serve_stop
      return 1
    fi

    id=${record%%$'\x1f'*}
    [[ $id == $_omp_serve_cycle ]] || continue
    payload=${record#*$'\x1f'}

    if [[ $payload == $'\x1e'* ]]; then
      _omp_transient_prompt=${payload#$'\x1e'}
      continue
    fi

    _omp_primary_prompt=$payload
    PS1=$payload
    return 0
  done
}

# async watcher: like _omp_async_handler, but for the daemon's id-prefixed
# records
function _omp_serve_async_handler() {
  local fd=$1
  local record

  IFS= read -r -u $fd -d $'\0' record 2>/dev/null
  if [[ $? -ne 0 ]]; then
    if [[ -z "$record" ]]; then
      # EOF - the daemon died. Only tear down if this is still the active fd;
      # a stale watcher must not close the current daemon's pipe.
      if [[ $_omp_serve_fd_out -eq $fd ]]; then
        _omp_serve_stop
      else
        zle -F $fd 2>/dev/null
      fi
      return 0
    fi
  fi

  local id=${record%%$'\x1f'*}
  [[ $id == $_omp_serve_cycle ]] || return 0
  local payload=${record#*$'\x1f'}

  if [[ $payload == $'\x1e'* ]]; then
    _omp_transient_prompt=${payload#$'\x1e'}
    return 0
  fi

  _omp_primary_prompt=$payload
  PS1=$payload
  zle reset-prompt 2>/dev/null

  return 0
}

function _omp_serve_abort() {
  setopt localoptions localtraps
  trap '' PIPE
  [[ $_omp_serve_fd_in -ge 0 ]] && print -r -u $_omp_serve_fd_in -- '{"command":"abort"}' 2>/dev/null
}

function _omp_serve_quit() {
  setopt localoptions localtraps
  trap '' PIPE
  [[ $_omp_serve_fd_in -ge 0 ]] && print -r -u $_omp_serve_fd_in -- '{"command":"quit"}' 2>/dev/null
  _omp_serve_stop
}

# shell exit handler
function _omp_exit_handler() {
  _omp_serve_quit
  _omp_cleanup_stream
}

# register exit handler
zshexit_functions+=(_omp_exit_handler)

# sets _omp_millis instead of printing to avoid forking a subshell
function _omp_milliseconds() {
  if (( ${+epochtime} )); then
    # copy first: every expansion of epochtime reads the clock again,
    # so referencing it twice in one expression can straddle a second boundary
    local -a now=("${epochtime[@]}")
    _omp_millis=$((now[1] * 1000 + now[2] / 1000000))
    return
  fi

  # zsh/datetime is unavailable
  _omp_millis=$($_omp_executable get millis)
}

# percent-encode $1 into REPLY, byte-wise, keeping RFC 3986 unreserved characters literal
function _omp_urlencode() {
  emulate -L zsh
  setopt no_multibyte
  local str=$1 ch
  local -i i
  REPLY=''
  for (( i = 1; i <= ${#str}; i++ )); do
    ch=$str[i]
    if [[ $ch == [A-Za-z0-9._~-] ]]; then
      REPLY+=$ch
      continue
    fi
    printf -v ch '%%%02X' "'$ch"
    REPLY+=$ch
  done
}

function _omp_preexec() {
  if [[ $_omp_ftcs_marks == 1 ]]; then
    if [[ -n $1 ]]; then
      # advertise the command line via kitty's cmdline_url= extension
      local REPLY
      _omp_urlencode "$1"
      printf '\033]133;C;cmdline_url=%s\007' "$REPLY"
    else
      printf '\033]133;C\007'
    fi
  fi

  _omp_milliseconds
  _omp_start_time=$_omp_millis
}

function _omp_precmd() {
  _omp_status=$?
  _omp_pipestatus=(${pipestatus[@]})
  _omp_job_count=${#jobstates}
  _omp_stack_count=${#dirstack[@]}
  _omp_execution_time=-1
  _omp_no_status=true
  _omp_tooltip_command=''

  if [[ -n $_omp_start_time ]]; then
    _omp_milliseconds
    _omp_execution_time=$(($_omp_millis - $_omp_start_time))
    _omp_no_status=false
  fi

  if [[ ${_omp_pipestatus[-1]} != "$_omp_status" ]]; then
    _omp_pipestatus=("$_omp_status")
  fi

  set_poshcontext
  _omp_set_cursor_position

  # We do this to avoid unexpected expansions in a prompt string.
  unsetopt PROMPT_SUBST
  unsetopt PROMPT_BANG

  # Ensure that escape sequences work in a prompt string.
  setopt PROMPT_PERCENT

  PS2=$_omp_secondary_prompt

  # === STREAMING PATH ===
  if [[ $_omp_enable_streaming -eq 1 ]]; then
    # set RPROMPT synchronously (records only carry the primary prompt)
    RPROMPT=$(_omp_get_prompt right)

    # serve daemon: persistent process, no spawn per prompt. After three
    # failures the daemon is left alone for the session and every prompt
    # takes the per-prompt stream below instead.
    if [[ $_omp_serve_failures -lt 3 ]] && _omp_serve_render; then
      unset _omp_start_time
      return 0
    fi

    # per-prompt stream (also the serve fallback; blocks until first prompt arrives)
    if _omp_start_streaming; then
      # PS1 already set by _omp_start_streaming (blocking first read)
      zle -F $_omp_stream_fd _omp_async_handler
      unset _omp_start_time
      return 0
    fi
    # fall through to sync
  fi

  # === FALLBACK PATH ===
  eval "$(_omp_get_prompt primary --eval)"

  unset _omp_start_time
}

# add hook functions
autoload -Uz add-zsh-hook
add-zsh-hook precmd _omp_precmd
add-zsh-hook preexec _omp_preexec

# Prevent incorrect behaviors when the initialization is executed twice in current session.
function _omp_cleanup() {
  local omp_widgets=(
    self-insert
    zle-line-init
    backward-delete-char
  )
  local widget
  for widget in "${omp_widgets[@]}"; do
    if [[ ${widgets[._omp_original::$widget]} ]]; then
      # Restore the original widget.
      zle -A ._omp_original::$widget $widget
    elif [[ ${widgets[$widget]} = user:_omp_* ]]; then
      # Delete the OMP-defined widget.
      zle -D $widget
    fi
  done
}
_omp_cleanup
unset -f _omp_cleanup

function _omp_get_prompt() {
  local type=$1
  local args=("${@[2,-1]}")
  $_omp_executable print $type \
    --save-cache \
    --shell=zsh \
    --shell-version=$ZSH_VERSION \
    --status=$_omp_status \
    --pipestatus="${_omp_pipestatus[*]}" \
    --no-status=$_omp_no_status \
    --execution-time=$_omp_execution_time \
    --job-count=$_omp_job_count \
    --stack-count=$_omp_stack_count \
    --terminal-width="${COLUMNS-0}" \
    ${args[@]}
}

function _omp_render_tooltip() {
  if [[ $KEYS != ' ' ]]; then
    return
  fi

  setopt local_options no_shwordsplit

  # Get the first word of command line as tip.
  local tooltip_command=${${(MS)BUFFER##[[:graph:]]*}%%[[:space:]]*}

  # Ignore an empty/repeated tooltip command.
  if [[ -z $tooltip_command ]] || [[ $tooltip_command = "$_omp_tooltip_command" ]]; then
    return
  fi

  _omp_tooltip_command="$tooltip_command"
  local tooltip=$(_omp_get_prompt tooltip --command="$tooltip_command")
  if [[ -z $tooltip ]]; then
    return
  fi

  RPROMPT=$tooltip
  zle .reset-prompt
}

function _omp_restore_rprompt() {
  if [[ -z $_omp_tooltip_command ]]; then
    return
  fi

  setopt local_options no_shwordsplit

  local current_command=${${(MS)BUFFER##[[:graph:]]*}%%[[:space:]]*}

  if [[ $current_command = "$_omp_tooltip_command" ]]; then
    return
  fi

  _omp_tooltip_command="$current_command"
  RPROMPT=$(_omp_get_prompt tooltip --command="$current_command")
  zle .reset-prompt
}

# Returns 0 when the current buffer is a complete command, 1 when more input
# is expected (used to keep the primary prompt while a multi-line command is
# still being typed).
function _omp_is_buffer_complete() {
  local buf=$PREBUFFER$BUFFER

  # 1. Syntax check. The buffer travels over a pipe rather than a here-document,
  # so no delimiter word can collide with whatever the user happens to type.
  if ! print -r -- "$buf" | zsh -n 2>/dev/null; then
    return 1
  fi

  # 2. An unterminated here-document parses cleanly, because the parser simply
  # ends the body at EOF. Step 1 therefore cannot see one. Re-parse with `;;`
  # appended: outside a case branch that is a syntax error, but an open
  # here-document swallows it as body text. A clean parse means the body was
  # never closed. The `<<` test only skips the extra parse when no here-document
  # can possibly be open.
  if [[ $buf == *'<<'* ]] && print -r -- "$buf"$'\n;;' | zsh -n 2>/dev/null; then
    return 1
  fi

  # 3. The SHORT_LOOPS option (on by default) makes `zsh -n` accept a bare loop
  # header as a complete empty loop - `for i in 1 2`, `while true`, or the
  # `for i in 1 2 do;` form - even though the line editor keeps waiting for the
  # body. Re-parse with an explicit `do : done` appended and SHORT_LOOPS off:
  # an open `for`/`while`/`until`/`select`/`repeat` header absorbs it into a
  # valid loop, while every genuinely complete command - including the
  # `for x (...) cmd` and `repeat n cmd` short forms - orphans the `do` into a
  # syntax error. The keyword test keeps the extra parse off the common path.
  if [[ $buf == *(for|while|until|select|repeat)* ]] && \
     print -r -- "$buf"$'\ndo\n:\ndone' | zsh +o shortloops -n 2>/dev/null; then
    return 1
  fi

  # 4. An odd number of trailing backslashes escapes the newline that follows the
  # buffer, which parses as a line continuation rather than an incomplete command.
  local trailing=${buf##*[^\\]}
  (( ${#trailing} % 2 )) && return 1

  return 0
}

function _omp_zle-line-init() {
  [[ $CONTEXT == start ]] || return 0

  # zsh-vi-mode wraps this widget, so its own line-init runs after ours - that
  # is, after .recursive-edit below has consumed the entire editing session.
  # Run it up front to align the keymap and ZVM's mode bookkeeping before
  # editing starts. See https://github.com/JanDeDobbeleer/oh-my-posh/issues/5992
  # The empty rawfunc shadows the like-named local in zvm_widget_wrapper:
  # zvm_reset_prompt resolves rawfunc dynamically and would otherwise re-enter
  # this widget through it.
  local rawfunc=
  if (( $+functions[zvm_zle-line-init] )) && [[ $ZVM_INIT_DONE == true ]]; then
    zvm_zle-line-init
  fi

  # Start regular line editor.
  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[1]

  local -i ret=0
  if [[ $POSH_MULTILINE_KEEPPROMPT == "true" ]]; then
    # Keep the full primary prompt while a multi-line command is still being
    # typed: re-enter the editor until the buffer is a complete command.
    while true; do
      zle .recursive-edit
      ret=$?

      if ((ret)) || _omp_is_buffer_complete; then
        break
      fi

      # Nothing may be prefixed to the continuation line. BUFFER is the command
      # itself, so anything added here is executed. Even pure whitespace breaks a
      # here-document, whose terminator only matches at the very start of a line.
      BUFFER+=$'\n'
      CURSOR=$#BUFFER
    done
  else
    zle .recursive-edit
    ret=$?
  fi

  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[2]

  # We need this workaround because when the `filler` is set,
  # there will be a redundant blank line below the transient prompt if the input is empty.
  local terminal_width_option
  if [[ -z $BUFFER ]]; then
    terminal_width_option="--terminal-width=$((${COLUMNS-0} - 1))"
  fi

  # stop prompt updates before the transient prompt renders so no handler
  # overwrites it: kill the per-prompt stream, tell the daemon to abort its
  # in-flight cycle (the daemon itself lives on)
  _omp_cleanup_stream
  _omp_serve_abort

  if ((ret)); then
    # interrupted (e.g. Ctrl-C): a pre-rendered transient prompt was built
    # before the interrupt and can't carry .Interrupted, so always re-render
    # through the CLI with the flag set
    eval "$(_omp_get_prompt transient --eval $terminal_width_option --interrupted)"
  elif [[ -n $_omp_transient_prompt ]]; then
    # rendered ahead of time by the streaming process (one column narrower,
    # mirroring the empty-buffer workaround above), saves a CLI call
    PS1=$_omp_transient_prompt
    RPROMPT=''
  else
    eval "$(_omp_get_prompt transient --eval $terminal_width_option)"
  fi
  zle .reset-prompt

  if ((ret)); then
    # TODO (fix): this is not equal to sending a SIGINT, since the status code ($?) is set to 1 instead of 130.
    zle .send-break
  fi

  # Exit the shell if we receive EOT.
  if [[ $KEYS == $'\4' ]]; then
    exit
  fi

  zle .accept-line
  return $ret
}

# Helper function for calling a widget before the specified OMP function.
function _omp_call_widget() {
  # The name of the OMP function.
  local omp_func=$1
  # The remainder are the widget to call and potential arguments.
  shift

  zle "$@" && shift 2 && $omp_func "$@"
}

# Create a widget with the specified OMP function.
# An existing widget will be preserved and decorated with the function.
function _omp_create_widget() {
  # The name of the widget to create/decorate.
  local widget=$1
  # The name of the OMP function.
  local omp_func=$2

  case ${widgets[$widget]:-''} in
  # Already decorated: do nothing.
  user:_omp_decorated_*) ;;

  # Non-existent: just create it.
  '')
    zle -N $widget $omp_func
    ;;

  # User-defined or builtin: backup and decorate it.
  *)
    # Back up the original widget. The leading dot in widget name is to work around bugs when used with zsh-syntax-highlighting in Zsh v5.8 or lower.
    zle -A $widget ._omp_original::$widget
    eval "_omp_decorated_${(q)widget}() { _omp_call_widget ${(q)omp_func} ._omp_original::${(q)widget} -- \"\$@\" }"
    zle -N $widget _omp_decorated_$widget
    ;;
  esac
}

function enable_poshtooltips() {
  local widget=${$(bindkey ' '):2}

  if [[ -z $widget ]]; then
    widget=self-insert
  fi

  _omp_create_widget $widget _omp_render_tooltip
  _omp_create_widget backward-delete-char _omp_restore_rprompt
}

# vi mode tracking — re-render the prompt whenever the active keymap changes
function _omp_render_vimode() {
  export POSH_VI_MODE=${KEYMAP:-main}
  eval "$(_omp_get_prompt primary --eval)"
  zle .reset-prompt
}

function _omp_enable_vimode() {
  export POSH_VI_MODE=${POSH_VI_MODE:-main}
  _omp_create_widget zle-keymap-select _omp_render_vimode
}

# This can be called by the user whenever re-rendering is required.
function omp_repaint_prompt() {
  eval "$(_omp_get_prompt primary --eval)"
  zle .reset-prompt 2>/dev/null
}

# Allow direct key binding: bindkey '^B' omp_repaint_prompt
zle -N omp_repaint_prompt

# legacy functions
function enable_poshtransientprompt() {}
