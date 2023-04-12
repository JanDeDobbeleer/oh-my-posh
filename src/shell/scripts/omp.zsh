export POSH_THEME=::CONFIG::
export POSH_SHELL_VERSION=$ZSH_VERSION
export POSH_PID=$$
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false
export POSH_PROMPT_COUNT=0

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
  echo -en "\033[6n" > /dev/tty
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
  omp_start_time=$(::OMP:: get millis)
}

function prompt_ohmyposh_precmd() {
  omp_last_error=$?
  omp_stack_count=${#dirstack[@]}
  omp_elapsed=-1
  if [ $omp_start_time ]; then
    local omp_now=$(::OMP:: get millis --shell=zsh)
    omp_elapsed=$(($omp_now-$omp_start_time))
  fi
  count=$((POSH_PROMPT_COUNT+1))
  export POSH_PROMPT_COUNT=$count
  set_poshcontext
  _set_posh_cursor_position
  eval "$(::OMP:: print primary --config="$POSH_THEME" --error="$omp_last_error" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION")"
  unset omp_start_time
}

# add hook functions
autoload -Uz add-zsh-hook
add-zsh-hook precmd prompt_ohmyposh_precmd
add-zsh-hook preexec prompt_ohmyposh_preexec

# perform cleanup so a new initialization in current session works
if [[ "$(zle -lL self-insert)" = *"_posh-tooltip"* ]]; then
  zle -N self-insert
fi
if [[ "$(zle -lL zle-line-init)" = *"_posh-zle-line-init"* ]]; then
  zle -N zle-line-init
fi

function _posh-tooltip() {
  # ignore an empty buffer
  if [[ -z  "$BUFFER"  ]]; then
    zle .self-insert
    return
  fi

  local tooltip=$(::OMP:: print tooltip --config="$POSH_THEME" --shell=zsh --error="$omp_last_error" --command="$BUFFER" --shell-version="$ZSH_VERSION")
  # ignore an empty tooltip
  if [[ -n "$tooltip" ]]; then
    RPROMPT=$tooltip
    zle .reset-prompt
  fi

  # https://github.com/zsh-users/zsh-autosuggestions - clear suggestion to avoid keeping it after the newly inserted space
  if [[ -n "$(zle -lL autosuggest-clear)" ]]; then
    # only if suggestions not disabled (variable not set)
    if ! [[ -v _ZSH_AUTOSUGGEST_DISABLED ]]; then
      zle autosuggest-clear
    fi
  fi
  zle .self-insert
  # https://github.com/zsh-users/zsh-autosuggestions - fetch new suggestion after the space
  if [[ -n "$(zle -lL autosuggest-fetch)" ]]; then
    # only if suggestions not disabled (variable not set)
    if ! [[ -v _ZSH_AUTOSUGGEST_DISABLED ]]; then
      zle autosuggest-fetch
    fi
  fi
}

function _posh-zle-line-init() {
    [[ $CONTEXT == start ]] || return 0

    # Start regular line editor
    (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[1]
    zle .recursive-edit
    local -i ret=$?
    (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[2]

    eval "$(::OMP:: print transient --error="$omp_last_error" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --config="$POSH_THEME" --eval --shell=zsh --shell-version="$ZSH_VERSION")"
    zle .reset-prompt

    # If we received EOT, we exit the shell
    if [[ $ret == 0 && $KEYS == $'\4' ]]; then
        exit
    fi

    # Ctrl-C
    if (( ret )); then
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

function enable_poshtransientprompt() {
  zle -N zle-line-init _posh-zle-line-init

  # restore broken key bindings
  # https://github.com/JanDeDobbeleer/oh-my-posh/discussions/2617#discussioncomment-3911044
  bindkey '^[[F' end-of-line
  bindkey '^[[H' beginning-of-line
  _widgets=$(zle -la)
  if [[ -n "${_widgets[(r)down-line-or-beginning-search]}" ]]; then
    bindkey '^[[B' down-line-or-beginning-search
  fi
  if [[ -n "${_widgets[(r)up-line-or-beginning-search]}" ]]; then
    bindkey '^[[A' up-line-or-beginning-search
  fi
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
