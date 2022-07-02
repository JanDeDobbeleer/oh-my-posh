set --export POSH_THEME '::CONFIG::'
set --global POWERLINE_COMMAND "oh-my-posh"
set --global CONDA_PROMPT_MODIFIER false
set --global omp_tooltip_command ""
set --global omp_transient 0

function fish_prompt
    # clear from cursor to end of screen as
    # commandline --function repaint does not do this
    # see https://github.com/fish-shell/fish-shell/issues/8418
    printf \e\[0J
    set --local omp_status_cache_temp $status
    if test "$omp_transient" = "1"
      '::OMP::' print transient --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
      return
    end
    set --global omp_status_cache $omp_status_cache_temp
    set --global omp_stack_count (count $dirstack)
    set --global omp_duration "$CMD_DURATION$cmd_duration"
    # check if variable set, < 3.2 case
    if set --query omp_lastcommand; and test "$omp_lastcommand" = ""
      set omp_duration 0
    end
    # works with fish >=3.2
    if set --query omp_last_status_generation; and test "$omp_last_status_generation" = "$status_generation"
      set omp_duration 0
    end
    if set --query status_generation
      set --global --export omp_last_status_generation $status_generation
    end

    '::OMP::' print primary --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
end

function fish_right_prompt
    if test "$omp_transient" = "1"
      echo -n ""
      set omp_transient 0
      return
    end
    if test -n "$omp_tooltip_command"
      set omp_tooltip_prompt ('::OMP::' print tooltip --config $POSH_THEME --shell fish --shell-version $FISH_VERSION --command $omp_tooltip_command)
      if test -n "$omp_tooltip_prompt"
        echo -n $omp_tooltip_prompt
        set omp_tooltip_command ""
        return
      end
    end
    '::OMP::' print right --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
end

function postexec_omp --on-event fish_postexec
  # works with fish <3.2
  # pre and postexec not fired for empty command in fish >=3.2
  set --global --export omp_lastcommand $argv
end

# tooltip

function _render_tooltip
  commandline --function expand-abbr
  set omp_tooltip_command (commandline --current-buffer | string collect)
  commandline --insert " "
  commandline --function repaint
end

function enable_poshtooltips
  bind \x20 _render_tooltip
end

# transient prompt

function _render_transient
  set omp_transient 1
  commandline --function repaint
  commandline --function execute
end

function enable_poshtransientprompt
  bind \r _render_transient
end
