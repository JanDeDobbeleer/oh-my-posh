set -g POSH_THEME "::CONFIG::"
set -g POWERLINE_COMMAND "oh-my-posh"
set -g CONDA_PROMPT_MODIFIER false

function fish_prompt
    set -g omp_status_cache $status
    set -g omp_stack_count (count $dirstack)
    set -g omp_duration "$CMD_DURATION$cmd_duration"
    # check if variable set, < 3.2 case
    if set -q omp_lastcommand; and test "$omp_lastcommand" = ""
      set omp_duration 0
    end
    # works with fish >=3.2
    if set -q omp_last_status_generation; and test "$omp_last_status_generation" = "$status_generation"
      set omp_duration 0
    end
    if set -q status_generation
      set -gx omp_last_status_generation $status_generation
    end

    ::OMP:: --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count
end

function fish_right_prompt
    ::OMP:: --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --rprompt
end

function postexec_omp --on-event fish_postexec
  # works with fish <3.2
  # pre and postexec not fired for empty command in fish >=3.2
  set -gx omp_lastcommand $argv
end


function export_poshconfig
  set -l file_name $argv[1]
  set -l format $argv[2]
  if not test -n "$file_name"
    echo "Usage: export_poshconfig \"filename\""
    return
  end
  if not test -n "$format"
    set format "json"
  end
  ::OMP:: --config $POSH_THEME --print-config --format $format > $file_name
end
