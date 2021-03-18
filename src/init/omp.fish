function fish_prompt
    set -l omp_duration "$CMD_DURATION$cmd_duration"
    if test "$omp_lastcommand" = ""
        set omp_duration 0
    end
    ::OMP:: --config ::CONFIG:: --error $status --execution-time $omp_duration
end

function postexec_omp --on-event fish_postexec
  set -gx omp_lastcommand $argv
end
