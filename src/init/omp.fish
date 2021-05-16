set -g posh_theme ::CONFIG::
set -g POWERLINE_COMMAND "oh-my-posh"

function fish_prompt
    set -l omp_stack_count (count $dirstack)
    set -l omp_duration "$CMD_DURATION$cmd_duration"
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

    ::OMP:: --config $posh_theme --error $status --execution-time $omp_duration --stack-count $omp_stack_count
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
  ::OMP:: --config $posh_theme --print-config --config-format $format > $file_name
end
