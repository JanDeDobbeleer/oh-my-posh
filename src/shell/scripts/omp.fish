set --export POSH_THEME ::CONFIG::
set --export POSH_SHELL_VERSION $FISH_VERSION
set --global POWERLINE_COMMAND "oh-my-posh"
set --global POSH_PID $fish_pid
set --global CONDA_PROMPT_MODIFIER false
set --global omp_tooltip_prompt ""
set --global has_omp_tooltip false
set --global omp_transient 0

# template function for context loading
function set_poshcontext
  return
end

function fish_prompt
    set --local omp_status_cache_temp $status
    # clear from cursor to end of screen as
    # commandline --function repaint does not do this
    # see https://github.com/fish-shell/fish-shell/issues/8418
    printf \e\[0J
    if test "$omp_transient" = "1"
      ::OMP:: print transient --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
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
    set_poshcontext
    # validate if the user cleared the screen
    set --local omp_cleared false
    set --local last_command (history search --max 1)
    if test "$last_command" = "clear"
      set omp_cleared true
    end
    ::OMP:: print primary --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION --cleared=$omp_cleared
end

function fish_right_prompt
    if test "$omp_transient" = "1"
      echo -n ""
      set omp_transient 0
      set has_omp_tooltip false
      return
    end
    if test -n "$omp_tooltip_prompt"
      echo -n $omp_tooltip_prompt
      set omp_tooltip_prompt  ""
      set has_omp_tooltip true
      return
    end
    set has_omp_tooltip false
    ::OMP:: print right --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
end

function postexec_omp --on-event fish_postexec
  # works with fish <3.2
  # pre and postexec not fired for empty command in fish >=3.2
  set --global --export omp_lastcommand $argv
end

# fix tooltip not resetting on SIGINT (ctrl+c)
function sigint_omp --on-signal INT
    commandline --function repaint
end

# perform cleanup so a new initialization in current session works
if test "$(string match -e '_render_transient' $(bind \r --user 2>/dev/null))" != ''
  bind -e \r
end
if test "$(string match -e '_render_tooltip' $(bind \x20 --user 2>/dev/null))" != ''
  bind -e \x20
end

# tooltip

function _render_tooltip
  commandline --function expand-abbr
  set omp_tooltip_command (commandline --current-buffer | string split --allow-empty -f1 ' ' | string collect)
  if not test -n "$omp_tooltip_command"
    return
  end
  set omp_tooltip_prompt (::OMP:: print tooltip --config $POSH_THEME --shell fish --error $omp_status_cache --shell-version $FISH_VERSION --command $omp_tooltip_command)
  commandline --insert " "
  if not test -n "$omp_tooltip_prompt"
    if test "$has_omp_tooltip" = "true"
      commandline --function repaint
    end
    return
  end
  commandline --function repaint
end

if test "::TOOLTIPS::" = "true"
  bind \x20 _render_tooltip -M default
  bind \x20 _render_tooltip -M insert
end

# transient prompt

function _render_transient
  if commandline --paging-mode
    commandline --function accept-autosuggestion
    return
  end
  set omp_transient 1
  commandline --function repaint
  commandline --function execute
end

if test "::TRANSIENT::" = "true"
  bind \r _render_transient -M default
  bind \r _render_transient -M insert
  bind \r _render_transient -M visual
end

# legacy functions
function enable_poshtooltips
  return
end
function enable_poshtransientprompt
  return
end

if test "::UPGRADE::" = "true"
  echo "::UPGRADENOTICE::"
end
