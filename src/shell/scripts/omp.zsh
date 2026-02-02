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
_omp_config=::CONFIG::
_omp_tooltip_command=''

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# set secondary prompt
_omp_secondary_prompt=''

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

function _omp_preexec() {
  if [[ $_omp_ftcs_marks == 1 ]]; then
    printf '\033]133;C\007'
  fi

  _omp_start_time=$($_omp_executable get millis)
}

function _omp_precmd() {
  if [[ -z $_omp_secondary_prompt ]]; then
    _omp_secondary_prompt=$($_omp_executable print secondary --shell=zsh)
  fi

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

  if [[ $_omp_daemon_mode == 1 ]] && [[ -n $_omp_transient_prompt ]]; then
    # Use daemon-provided transient prompt
    PS1=$_omp_transient_prompt
  else
    # We need this workaround because when the `filler` is set,
    # there will be a redundant blank line below the transient prompt if the input is empty.
    local terminal_width_option
    if [[ -z $BUFFER ]]; then
      terminal_width_option="--terminal-width=$((${COLUMNS-0} - 1))"
    fi
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

# Daemon mode variables
_omp_daemon_mode=0
_omp_daemon_fd=
_omp_transient_prompt=

function _omp_daemon_precmd() {
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

  unsetopt PROMPT_SUBST
  unsetopt PROMPT_BANG
  setopt PROMPT_PERCENT

  PS2=$_omp_secondary_prompt

  # Clean up any existing fd handler from previous prompt
  if [[ -n $_omp_daemon_fd ]]; then
    zle -F $_omp_daemon_fd 2>/dev/null
    exec {_omp_daemon_fd}<&- 2>/dev/null
    _omp_daemon_fd=
  fi

  # Start daemon render process
  local fd
  exec {fd}< <($_omp_executable render \
    --config=$_omp_config \
    --shell=zsh \
    --shell-version=$ZSH_VERSION \
    --pwd="$PWD" \
    --pid=$$ \
    --status=$_omp_status \
    --pipestatus="${_omp_pipestatus[*]}" \
    --no-status=$_omp_no_status \
    --execution-time=$_omp_execution_time \
    --job-count=$_omp_job_count \
    --stack-count=$_omp_stack_count \
    --terminal-width="${COLUMNS-0}" \
    2>/dev/null)

  # Read first batch synchronously (partial results after daemon timeout)
  local line batch_complete=0
  while [[ $batch_complete -eq 0 ]] && IFS= read -r line <&$fd; do
    _omp_daemon_parse_line "$line"
    if [[ $line == status:* ]]; then
      batch_complete=1
      if [[ $line == "status:complete" ]]; then
        # All done, close fd
        exec {fd}<&-
        unset _omp_start_time
        return
      fi
    fi
  done

  # More updates may come - register fd handler for streaming
  _omp_daemon_fd=$fd
  zle -F $fd _omp_daemon_handler

  unset _omp_start_time
}

function _omp_daemon_parse_line() {
  local line=$1
  local type=${line%%:*}
  local text=${line#*:}

  case $type in
    primary)
      PS1=$text
      ;;
    right)
      RPROMPT=$text
      ;;
    secondary)
      PS2=$text
      ;;
    transient)
      _omp_transient_prompt=$text
      ;;
  esac
}

function _omp_daemon_handler() {
  local fd=$1
  local line batch_complete=0

  # Read all available lines in this batch
  while [[ $batch_complete -eq 0 ]] && IFS= read -r -t 0 line <&$fd; do
    _omp_daemon_parse_line "$line"
    if [[ $line == status:* ]]; then
      batch_complete=1
    fi
  done

  # If we read at least one status line, repaint
  if [[ $batch_complete -eq 1 ]]; then
    zle .reset-prompt
  fi

  # Check if stream ended
  if [[ $line == "status:complete" ]] || ! IFS= read -r -t 0.01 line <&$fd 2>/dev/null; then
    # Check if fd is still valid by trying to read
    if ! IFS= read -r -t 0.01 _ <&$fd 2>/dev/null; then
      zle -F $fd 2>/dev/null
      exec {fd}<&- 2>/dev/null
      _omp_daemon_fd=
    fi
  fi
}

function enable_poshdaemon() {
  # Start daemon if not running
  $_omp_executable daemon start --config=$_omp_config --silent >/dev/null 2>&1 &!

  # Replace precmd with daemon version
  _omp_daemon_mode=1
  add-zsh-hook -d precmd _omp_precmd
  add-zsh-hook precmd _omp_daemon_precmd
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
