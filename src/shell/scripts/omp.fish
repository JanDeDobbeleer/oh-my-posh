set --export POSH_THEME ::CONFIG::
set --global POWERLINE_COMMAND "oh-my-posh"
set --global CONDA_PROMPT_MODIFIER false
set --global omp_tooltip_command ""
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
    ::OMP:: print primary --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
end

function fish_right_prompt
    if test "$omp_transient" = "1"
      echo -n ""
      set omp_transient 0
      return
    end
    if test -n "$omp_tooltip_command"
      set omp_tooltip_prompt (::OMP:: print tooltip --config $POSH_THEME --shell fish --error $omp_status_cache --shell-version $FISH_VERSION --command $omp_tooltip_command)
      if test -n "$omp_tooltip_prompt"
        echo -n $omp_tooltip_prompt
        set omp_tooltip_command ""
        return
      end
    end
    ::OMP:: print right --config $POSH_THEME --shell fish --error $omp_status_cache --execution-time $omp_duration --stack-count $omp_stack_count --shell-version $FISH_VERSION
end

function postexec_omp --on-event fish_postexec
  # works with fish <3.2
  # pre and postexec not fired for empty command in fish >=3.2
  set --global --export omp_lastcommand $argv
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
  set omp_tooltip_command (commandline --current-buffer | string collect | string split --allow-empty -f1 ' ')
  commandline --insert " "
  commandline --function repaint
end

if test "::TOOLTIPS::" = "true"
  bind \x20 _render_tooltip
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
  bind \r _render_transient
end

# legacy functions
function enable_poshtooltips
  return
end
function enable_poshtransientprompt
  return
end
