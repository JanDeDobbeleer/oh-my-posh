function fish_prompt
    set -l omp_duration "$CMD_DURATION$cmd_duration"
    ::OMP:: --config ::CONFIG:: --error $status --execution-time $omp_duration
end
