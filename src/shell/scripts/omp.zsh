export POSH_THEME=::CONFIG::
export POSH_SHELL_VERSION=$ZSH_VERSION
export POSH_PID=$$
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false
export POSH_PROMPT_COUNT=0
export ZLE_RPROMPT_INDENT=0

_omp_executable=::OMP::

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# set secondary prompt
PS2="$(${_omp_executable} print secondary --config="$POSH_THEME" --shell=zsh)"

function _omp_set_cursor_position() {
  # not supported in Midnight Commander
  # see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3415
  if [[ $_omp_cursor_positioning == 0 ]] || [[ -v MC_SID ]]; then
    return
  fi

  local oldstty=$(stty -g)
  stty raw -echo min 0

  local pos
  echo -en "\033[6n" >/dev/tty
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
  if [[ $_omp_ftcs_marks == 0 ]]; then
    printf "\033]133;C\007"
  fi

  _omp_start_time=$(${_omp_executable} get millis)
}

function _omp_precmd() {
  _omp_status_cache=$?
  _omp_pipestatus_cache=(${pipestatus[@]})
  _omp_stack_count=${#dirstack[@]}
  _omp_elapsed=-1
  _omp_no_exit_code="true"

  if [ $_omp_start_time ]; then
    local omp_now=$(${_omp_executable} get millis --shell=zsh)
    _omp_elapsed=$(($omp_now - $_omp_start_time))
    _omp_no_exit_code="false"
  fi

  if [[ "${_omp_pipestatus_cache[-1]}" != "$_omp_status_cache" ]]; then
    _omp_pipestatus_cache=("$_omp_status_cache")
  fi

  count=$((POSH_PROMPT_COUNT + 1))
  export POSH_PROMPT_COUNT=$count

  set_poshcontext
  _omp_set_cursor_position

  eval "$(${_omp_executable} print primary --config="$POSH_THEME" --status="$_omp_status_cache" --pipestatus="${_omp_pipestatus_cache[*]}" --execution-time="$_omp_elapsed" --stack-count="$_omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$_omp_no_exit_code")"
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
    if [[ "${widgets[_omp_original::$widget]}" ]]; then
      # Restore the original widget.
      zle -A _omp_original::$widget $widget
    elif [[ "${widgets[$widget]}" = user:_omp_* ]]; then
      # Delete the OMP-defined widget.
      zle -D $widget
    fi
  done
}
_omp_cleanup
unset -f _omp_cleanup

function _omp_render_tooltip() {
  if [[ "$KEYS" != ' ' ]]; then
    return
  fi

  # Get the first word of command line as tip.
  local tooltip_command=${${(MS)BUFFER##[[:graph:]]*}%%[[:space:]]*}

  # Ignore an empty/repeated tooltip command.
  if [[ -z "$tooltip_command" ]] || [[ "$tooltip_command" = "$_omp_tooltip_command" ]]; then
    return
  fi

  _omp_tooltip_command="$tooltip_command"
  local tooltip=$(${_omp_executable} print tooltip --config="$POSH_THEME" --status="$_omp_status_cache" --pipestatus="${_omp_pipestatus_cache[*]}" --execution-time="$_omp_elapsed" --stack-count="$_omp_stack_count" --command="$tooltip_command" --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$_omp_no_exit_code")
  if [[ -z "$tooltip" ]]; then
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

  _omp_tooltip_command=''
  eval "$(${_omp_executable} print transient --config="$POSH_THEME" --status="$_omp_status_cache" --pipestatus="${_omp_pipestatus_cache[*]}" --execution-time="$_omp_elapsed" --stack-count="$_omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$_omp_no_exit_code")"
  zle .reset-prompt

  # Exit the shell if we receive EOT.
  if [[ $ret == 0 && $KEYS == $'\4' ]]; then
    exit
  fi

  if ((ret)); then
    # TODO (fix): this is not equal to sending a SIGINT, since the status code ($?) is set to 1 instead of 130.
    zle .send-break
  else
    # Enter
    zle .accept-line
  fi
  return ret
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
    # Back up the original widget.
    zle -A $widget _omp_original::$widget
    eval "_omp_decorated_${(q)widget}() { _omp_call_widget ${(q)omp_func} _omp_original::${(q)widget} -- \"\$@\" }"
    zle -N $widget _omp_decorated_$widget
    ;;
  esac
}

# legacy functions
function enable_poshtooltips() {}
function enable_poshtransientprompt() {}
