export POSH_THEME=::CONFIG::
export POSH_SHELL='zsh'
export POSH_SHELL_VERSION=$ZSH_VERSION
export POSH_SESSION_ID=::SESSION_ID::
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
    ${args[@]}
}

function _omp_render_tooltip() {
  if [[ $KEYS != ' ' ]]; then
    return
  fi

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

  eval "$(_omp_get_prompt transient --eval)"
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
