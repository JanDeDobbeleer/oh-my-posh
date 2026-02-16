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

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# streaming support variables
_omp_stream_fd=-1
_omp_enable_streaming=0
_omp_primary_prompt=""
_omp_streaming_supported=""

# set secondary prompt
_omp_secondary_prompt=$($_omp_executable print secondary --shell=zsh)

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
  exec {_omp_stream_fd}< <(exec "${stream_cmd[@]}") 2>/dev/null

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

  # read single null-delimited prompt (stream emits one value per \0)
  IFS= read -r -u $fd -d $'\0' _omp_primary_prompt
  if [[ $? -ne 0 ]]; then
    if [[ -z "$_omp_primary_prompt" ]]; then
      # EOF — stream closed normally
      _omp_cleanup_stream
      return 0
    fi
  fi

  PS1="$_omp_primary_prompt"
  zle reset-prompt 2>/dev/null

  return 0
}

# shell exit handler
function _omp_exit_handler() {
  _omp_cleanup_stream
}

# register exit handler
zshexit_functions+=(_omp_exit_handler)

function _omp_preexec() {
  if [[ $_omp_ftcs_marks == 1 ]]; then
    printf '\033]133;C\007'
  fi

  _omp_start_time=$($_omp_executable get millis)
}

function _omp_precmd() {
  _omp_status=$?
  _omp_pipestatus=(${pipestatus[@]})
  _omp_job_count=${#jobstates}
  _omp_stack_count=${#dirstack[@]}
  _omp_execution_time=-1
  _omp_no_status=true
  _omp_tooltip_command=''

  if [ $_omp_start_time ]; then
    local omp_now=$($_omp_executable get millis)
    _omp_execution_time=$(($omp_now - $_omp_start_time))
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
    # set RPROMPT synchronously (stream only emits primary prompt)
    RPROMPT=$(_omp_get_prompt right)

    # start new streaming session (blocks until first prompt arrives)
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

function _omp_zle-line-init() {
  [[ $CONTEXT == start ]] || return 0

  # Start regular line editor.
  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[1]
  zle .recursive-edit
  local -i ret=$?
  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[2]

  # We need this workaround because when the `filler` is set,
  # there will be a redundant blank line below the transient prompt if the input is empty.
  local terminal_width_option
  if [[ -z $BUFFER ]]; then
    terminal_width_option="--terminal-width=$((${COLUMNS-0} - 1))"
  fi

  # kill streaming before transient prompt to prevent handler overwriting it
  _omp_cleanup_stream

  eval "$(_omp_get_prompt transient --eval $terminal_width_option)"
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
}

# legacy functions
function enable_poshtransientprompt() {}
