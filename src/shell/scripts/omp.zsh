export POSH_THEME=::CONFIG::
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false

# set secondary prompt
PS2="$(::OMP:: print secondary --config="$POSH_THEME" --shell=zsh)"

function prompt_ohmyposh_preexec() {
  omp_start_time=$(::OMP:: get millis)
}

function prompt_ohmyposh_precmd() {
  omp_last_error=$?
  omp_stack_count=${#dirstack[@]}
  omp_elapsed=-1
  if [ $omp_start_time ]; then
    omp_now=$(::OMP:: get millis)
    omp_elapsed=$(($omp_now-$omp_start_time))
  fi
  eval "$(::OMP:: print primary --config="$POSH_THEME" --error="$omp_last_error" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --eval --shell=zsh --shell-version="$ZSH_VERSION")"
  unset omp_start_time
  unset omp_now
}

function _install-omp-hooks() {
  for s in "${preexec_functions[@]}"; do
    if [ "$s" = "prompt_ohmyposh_preexec" ]; then
      return
    fi
  done
  preexec_functions+=(prompt_ohmyposh_preexec)

  for s in "${precmd_functions[@]}"; do
    if [ "$s" = "prompt_ohmyposh_precmd" ]; then
      return
    fi
  done
  precmd_functions+=(prompt_ohmyposh_precmd)
}

if [ "$TERM" != "linux" ]; then
  _install-omp-hooks
fi

# perform cleanup so a new initialization in current session works
if [[ "$(zle -lL self-insert)" = *"_posh-self-insert"* ]]; then
  zle -N self-insert
fi
if [[ "$(zle -lL zle-line-init)" = *"_posh-zle-line-init"* ]]; then
  zle -N zle-line-init
fi

function _posh-self-insert() {
  # ignore an empty buffer
  if [[ -z  "$BUFFER"  ]]; then
    zle .self-insert
    return
  fi
  # trigger a tip check only if the input is a space character
  if [[ "$KEYS" = " " ]]; then
    local tooltip=$(::OMP:: print tooltip --config="$POSH_THEME" --shell=zsh --error="$omp_last_error" --command="$BUFFER" --shell-version="$ZSH_VERSION")
  fi
  # ignore an empty tooltip
  if [[ -n "$tooltip" ]]; then
    RPROMPT=$tooltip
    zle .reset-prompt
  fi
  zle .self-insert
}

if [[ "::TOOLTIPS::" = "true" ]]; then
  zle -N self-insert _posh-self-insert
fi

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

if [[ "::TRANSIENT::" = "true" ]]; then
  zle -N zle-line-init _posh-zle-line-init
fi

# legacy functions for backwards compatibility
function enable_poshtooltips() {}
function enable_poshtransientprompt() {}
