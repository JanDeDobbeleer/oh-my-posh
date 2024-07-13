export POSH_THEME=::CONFIG::
export POSH_SHELL_VERSION=$ZSH_VERSION
export POSH_PID=$$
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false
export POSH_PROMPT_COUNT=0
export ZLE_RPROMPT_INDENT=0

# set secondary prompt
PS2="$(::OMP:: print secondary --config="$POSH_THEME" --shell=zsh)"

function _set_posh_cursor_position() {
  # not supported in Midnight Commander
  # see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3415
  if [[ "::CURSOR::" != "true" ]] || [[ -v MC_SID ]]; then
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

function prompt_ohmyposh_preexec() {
  if [[ "::FTCS_MARKS::" = "true" ]]; then
    printf "\033]133;C\007"
  fi

  omp_start_time=$(::OMP:: get millis)
}

function prompt_ohmyposh_precmd() {
  omp_status_cache=$?
  omp_pipestatus_cache=(${pipestatus[@]})
  omp_stack_count=${#dirstack[@]}
  omp_elapsed=-1
  omp_no_exit_code="true"

  if [ $omp_start_time ]; then
    local omp_now=$(::OMP:: get millis --shell=zsh)
    omp_elapsed=$(($omp_now - $omp_start_time))
    omp_no_exit_code="false"
  fi

  if [[ "${omp_pipestatus_cache[-1]}" != "$omp_status_cache" ]]; then
    omp_pipestatus_cache=("$omp_status_cache")
  fi

  count=$((POSH_PROMPT_COUNT + 1))
  export POSH_PROMPT_COUNT=$count

  set_poshcontext
  _set_posh_cursor_position

  eval "$(::OMP:: print primary --config="$POSH_THEME" --status="$omp_status_cache" --pipestatus="${omp_pipestatus_cache[*]}" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$omp_no_exit_code")"
  unset omp_start_time
}

# add hook functions
autoload -Uz add-zsh-hook
add-zsh-hook precmd prompt_ohmyposh_precmd
add-zsh-hook preexec prompt_ohmyposh_preexec

# perform cleanup so a new initialization in current session works
if [[ "$(bindkey " ")" = *"_posh-tooltip"* ]]; then
  bindkey " " self-insert
fi
if [[ "$(zle -lL zle-line-init)" = *"_posh-zle-line-init"* ]]; then
  zle -N zle-line-init
fi

function _posh-tooltip() {
  # https://github.com/zsh-users/zsh-autosuggestions - clear suggestion to avoid keeping it after the newly inserted space
  if [[ "$(zle -lL autosuggest-clear)" ]]; then
    # only if suggestions not disabled (variable not set)
    if [[ ! -v _ZSH_AUTOSUGGEST_DISABLED ]]; then
      zle autosuggest-clear
    fi
  fi

  zle .self-insert

  # https://github.com/zsh-users/zsh-autosuggestions - fetch new suggestion after the space
  if [[ "$(zle -lL autosuggest-fetch)" ]]; then
    # only if suggestions not disabled (variable not set)
    if [[ ! -v _ZSH_AUTOSUGGEST_DISABLED ]]; then
      zle autosuggest-fetch
    fi
  fi

  # Get the first word of command line as tip.
  local tooltip_command=${${(MS)BUFFER##[[:graph:]]*}%%[[:space:]]*}

  # Ignore an empty/repeated tooltip command.
  if [[ -z "$tooltip_command" ]] || [[ "$tooltip_command" = "$omp_tooltip_command" ]]; then
    return
  fi

  omp_tooltip_command="$tooltip_command"
  local tooltip=$(::OMP:: print tooltip --config="$POSH_THEME" --status="$omp_status_cache" --pipestatus="${omp_pipestatus_cache[*]}" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --command="$tooltip_command" --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$omp_no_exit_code")
  if [[ -z "$tooltip" ]]; then
    return
  fi

  RPROMPT=$tooltip
  zle .reset-prompt
}

function _posh-zle-line-init() {
  [[ $CONTEXT == start ]] || return 0

  # Start regular line editor.
  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[1]
  zle .recursive-edit
  local -i ret=$?
  (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[2]

  omp_tooltip_command=''
  eval "$(::OMP:: print transient --config="$POSH_THEME" --status="$omp_status_cache" --pipestatus="${omp_pipestatus_cache[*]}" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION" --no-status="$omp_no_exit_code")"
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

function enable_poshtooltips() {
  zle -N _posh-tooltip
  bindkey " " _posh-tooltip
}

# Helper function for posh::decorate_widget
# It calls the posh function right after the original definition of the widget
# $1 is the posh widget name
# $2.. are the name of the widget to call + potential args
posh::call_widget()
{
  local posh_widget=$1;shift
  builtin zle "${@}" &&
  ${posh_widget}
}

# decorate_widget
# Allows to preserve any user defined value that may have been defined before posh tries to redefine it.
# Instead, we keep the previous function and decorate it with the posh additions
# $1: The name of the widget to decorate
# $2: The name of the posh function to decorate it with
function posh::decorate_widget() {
  orig_widget=${1};shift
  posh_widget=${1};shift
  # from this point $@ does not have $1 $2 anymore

  case ${widgets[$orig_widget]:-""} in
    # Already decorated: do nothing.
    user:_posh-decorated-*);;

    # User defined
    user:*)
      zle -N $POSH_PID-$orig_widget ${widgets[$orig_widget]#*:}
      eval "_posh-decorated-${(q)POSH_PID}-${(q)orig_widget}() { posh::call_widget ${(q)posh_widget} ${(q)POSH_PID}-${(q)orig_widget} -- \"\$@\" }"
      zle -N $orig_widget _posh-decorated-$POSH_PID-$orig_widget;;

    # Built-in
    builtin:*)
      eval "_posh-decorated-${(q)POSH_PID}-${(q)orig_widget}() { posh::call_widget ${(q)posh_widget} .${(q)orig_widget} -- \"\$@\" }"
      zle -N $orig_widget _posh-decorated-$POSH_PID-$orig_widget;;

    # non-existent
    *)
      if [[ $orig_widget == zle-* ]] && (( ! ${+widgets[$orig_widget]} )); then
        # The widget is a zle one and does not exist, we can safely create it
        # Otherwise, do nothing
        eval "_posh-decorated-${(q)POSH_PID}-${(q)orig_widget}() { ${(q)posh_widget} }"
        zle -N $orig_widget _posh-decorated-$POSH_PID-$orig_widget
      fi
      ;;
  esac
}

function enable_poshtransientprompt() {
  posh::decorate_widget zle-line-init _posh-zle-line-init
}

if [[ "::TOOLTIPS::" = "true" ]]; then
  enable_poshtooltips
fi

if [[ "::TRANSIENT::" = "true" ]]; then
  enable_poshtransientprompt
fi

if [[ "::UPGRADE::" = "true" ]]; then
  echo "::UPGRADENOTICE::"
fi

if [[ "::AUTOUPGRADE::" = "true" ]]; then
  ::OMP:: upgrade
fi
