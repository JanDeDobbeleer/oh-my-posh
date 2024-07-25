export POSH_THEME=::CONFIG::
export POSH_SHELL_VERSION=$BASH_VERSION
export POWERLINE_COMMAND="oh-my-posh"
export POSH_PID=$$
export CONDA_PROMPT_MODIFIER=false
export OSTYPE=$OSTYPE

# global variables
_omp_start_time=""
_omp_stack_count=0
_omp_elapsed=-1
_omp_no_exit_code="true"
_omp_status_cache=0
_omp_pipestatus_cache=0
_omp_executable=::OMP::

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# start timer on command start
PS0='${_omp_start_time:0:$((_omp_start_time="$(_omp_start_timer)",0))}$(_omp_ftcs_command_start)'
# set secondary prompt
PS2="$(${_omp_executable} print secondary --shell=bash --shell-version="$BASH_VERSION")"

function _omp_set_cursor_position() {
    # not supported in Midnight Commander
    # see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3415
    if [[ $_omp_cursor_positioning == 0 ]] || [[ -v MC_SID ]]; then
        return
    fi

    local oldstty=$(stty -g)
    stty raw -echo min 0

    local COL
    local ROW
    IFS=';' read -sdR -p $'\E[6n' ROW COL

    stty $oldstty

    export POSH_CURSOR_LINE=${ROW#*[}
    export POSH_CURSOR_COLUMN=${COL}
}

function _omp_start_timer() {
    $_omp_executable get millis
}

function _omp_ftcs_command_start() {
    if [[ $_omp_ftcs_marks == 0 ]]; then
        printf "\e]133;C\a"
    fi
}

# template function for context loading
function set_poshcontext() {
    return
}

# regular prompt
function _omp_hook() {
    _omp_status_cache=$? _omp_pipestatus_cache=(${PIPESTATUS[@]})

    if [[ "${#BP_PIPESTATUS[@]}" -ge "${#_omp_pipestatus_cache[@]}" ]]; then
        _omp_pipestatus_cache=(${BP_PIPESTATUS[@]})
    fi

    _omp_stack_count=$((${#DIRSTACK[@]} - 1))

    if [[ "$_omp_start_time" ]]; then
        local omp_now=$(${_omp_executable} get millis --shell=bash)
        _omp_elapsed=$((omp_now - _omp_start_time))
        _omp_start_time=""
        _omp_no_exit_code="false"
    fi

    if [[ "${_omp_pipestatus_cache[-1]}" != "$_omp_status_cache" ]]; then
        _omp_pipestatus_cache=("$_omp_status_cache")
    fi

    set_poshcontext
    _omp_set_cursor_position

    PS1="$(${_omp_executable} print primary --shell=bash --shell-version="$BASH_VERSION" --status="$_omp_status_cache" --pipestatus="${_omp_pipestatus_cache[*]}" --execution-time="$_omp_elapsed" --stack-count="$_omp_stack_count" --no-status="$_omp_no_exit_code" --terminal-width="${COLUMNS-0}" | tr -d '\0')"

    return $_omp_status_cache
}

if [[ "$TERM" != "linux" ]] && [[ -x "$(command -v $_omp_executable)" ]] && ! [[ "$PROMPT_COMMAND" =~ "_omp_hook" ]]; then
    PROMPT_COMMAND="_omp_hook; $PROMPT_COMMAND"
fi
