export POSH_THEME=::CONFIG::
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false

function omp_preexec() {
  omp_start_time=$(::OMP:: --millis)
}

function omp_precmd() {
  omp_last_error=$?
  omp_stack_count=${#dirstack[@]}
  omp_elapsed=-1
  if [ $omp_start_time ]; then
    omp_now=$(::OMP:: --millis)
    omp_elapsed=$(($omp_now-$omp_start_time))
  fi
  eval "$(::OMP:: --config $POSH_THEME --error $omp_last_error --execution-time $omp_elapsed --stack-count $omp_stack_count --eval --shell zsh)"
  unset omp_start_time
  unset omp_now
  unset omp_elapsed
  unset omp_last_error
  unset omp_stack_count
}

function install_omp_hooks() {
  for s in "${preexec_functions[@]}"; do
    if [ "$s" = "omp_preexec" ]; then
      return
    fi
  done
  preexec_functions+=(omp_preexec)

  for s in "${precmd_functions[@]}"; do
    if [ "$s" = "omp_precmd" ]; then
      return
    fi
  done
  precmd_functions+=(omp_precmd)
}

if [ "$TERM" != "linux" ]; then
  install_omp_hooks
fi

function export_poshconfig() {
    [ $# -eq 0 ] && { echo "Usage: $0 \"filename\""; return; }
    format=$2
    if [ -z "$format" ]; then
      format="json"
    fi
    ::OMP:: --config $POSH_THEME --print-config --config-format $format > $1
}

function self-insert() {
  # ignore an empty buffer
  if [[ -z  "$BUFFER"  ]]; then
    zle .self-insert
    return
  fi
  tooltip=$(::OMP:: --config $POSH_THEME --shell zsh --command $BUFFER)
  # ignore an empty tooltip
  if [[ ! -z "$tooltip" ]]; then
    RPROMPT=$tooltip
    zle reset-prompt
  fi
  zle .self-insert
}

function enable_poshtooltips() {
  zle -N self-insert
}
