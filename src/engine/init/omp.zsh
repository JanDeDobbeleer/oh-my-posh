export POSH_THEME="::CONFIG::"
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false

# set secondary prompt
PS2="$(::OMP:: prompt print secondary --config="$POSH_THEME" --shell=zsh)"

function prompt_ohmyposh_preexec() {
  omp_start_time=$(::OMP:: config get millis)
}

function prompt_ohmyposh_precmd() {
  omp_last_error=$?
  omp_stack_count=${#dirstack[@]}
  omp_elapsed=-1
  if [ $omp_start_time ]; then
    omp_now=$(::OMP:: config get millis)
    omp_elapsed=$(($omp_now-$omp_start_time))
  fi
  eval "$(::OMP:: prompt print primary --config="$POSH_THEME" --error="$omp_last_error" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" --eval --shell=zsh)"
  unset omp_start_time
  unset omp_now
  unset omp_elapsed
  unset omp_last_error
  unset omp_stack_count
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

function self-insert() {
  # ignore an empty buffer
  if [[ -z  "$BUFFER"  ]]; then
    zle .self-insert
    return
  fi
  tooltip=$(::OMP:: prompt print tooltip --config="$POSH_THEME" --shell=zsh --command="$BUFFER")
  # ignore an empty tooltip
  if [[ ! -z "$tooltip" ]]; then
    RPROMPT=$tooltip
    zle .reset-prompt
  fi
  zle .self-insert
}

function enable_poshtooltips() {
  zle -N self-insert
}

_posh-zle-line-init() {
    [[ $CONTEXT == start ]] || return 0

    # Start regular line editor
    (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[1]
    zle .recursive-edit
    local -i ret=$?
    (( $+zle_bracketed_paste )) && print -r -n - $zle_bracketed_paste[2]

    eval "$(::OMP:: prompt print transient --config="$POSH_THEME" --eval --shell=zsh)"
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

function enable_poshtransientprompt() {
  zle -N zle-line-init _posh-zle-line-init
}
