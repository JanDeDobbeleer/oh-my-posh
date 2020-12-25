export POSH_THEME=::CONFIG::

function omp_preexec() {
  omp_start_time=$(::OMP:: --millis)
}

function omp_precmd() {
  omp_elapsed=-1
  if [ $omp_start_time ]; then
    omp_now=$(::OMP:: --millis)
    omp_elapsed=$(($omp_now-$omp_start_time))
  fi
  eval "$(::OMP:: --config $POSH_THEME --error $? --execution-time $omp_elapsed --eval)"
  unset omp_start_time
  unset omp_now
  unset omp_elapsed
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
