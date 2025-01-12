export POSH_THEME=::CONFIG::
export POSH_SHELL='bash'
export POSH_SHELL_VERSION=$BASH_VERSION
export POWERLINE_COMMAND='oh-my-posh'
export POSH_SESSION_ID=::SESSION_ID::
export CONDA_PROMPT_MODIFIER=false
export OSTYPE=$OSTYPE

# disable all known python virtual environment prompts
export VIRTUAL_ENV_DISABLE_PROMPT=1
export PYENV_VIRTUALENV_DISABLE_PROMPT=1

# global variables
_omp_start_time=''
_omp_stack_count=0
_omp_execution_time=-1
_omp_no_status=true
_omp_status=0
_omp_pipestatus=0
_omp_executable=::OMP::

# switches to enable/disable features
_omp_cursor_positioning=0
_omp_ftcs_marks=0

# start timer on command start
PS0='${_omp_start_time:0:$((_omp_start_time="$(_omp_start_timer)",0))}$(_omp_ftcs_command_start)'

# set secondary prompt
_omp_secondary_prompt=$(
    "$_omp_executable" print secondary \
        --shell=bash \
        --shell-version="$BASH_VERSION"
)

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
    IFS=';' read -rsdR -p $'\E[6n' ROW COL

    stty "$oldstty"

    export POSH_CURSOR_LINE=${ROW#*[}
    export POSH_CURSOR_COLUMN=${COL}
}

function _omp_start_timer() {
    "$_omp_executable" get millis
}

function _omp_ftcs_command_start() {
    if [[ $_omp_ftcs_marks == 1 ]]; then
        printf '\e]133;C\a'
    fi
}

# template function for context loading
function set_poshcontext() {
    return
}

function _omp_get_primary() {
    # Avoid unexpected expansions when we're generating the prompt below.
    shopt -u promptvars
    trap 'shopt -s promptvars' RETURN

    local prompt
    if shopt -oq posix; then
        # Disable in POSIX mode.
        prompt='[NOTICE: Oh My Posh prompt is not supported in POSIX mode]\n\u@\h:\w\$ '
    else
        prompt=$(
            "$_omp_executable" print primary \
                --save-cache \
                --shell=bash \
                --shell-version="$BASH_VERSION" \
                --status="$_omp_status" \
                --pipestatus="${_omp_pipestatus[*]}" \
                --no-status="$_omp_no_status" \
                --execution-time="$_omp_execution_time" \
                --stack-count="$_omp_stack_count" \
                --terminal-width="${COLUMNS-0}" |
                tr -d '\0'
        )
    fi
    echo "${prompt@P}"
}

function _omp_get_secondary() {
    # Avoid unexpected expansions when we're generating the prompt below.
    shopt -u promptvars
    trap 'shopt -s promptvars' RETURN

    if shopt -oq posix; then
        # Disable in POSIX mode.
        echo '> '
    else
        echo "${_omp_secondary_prompt@P}"
    fi
}

function _omp_hook() {
    _omp_status=$? _omp_pipestatus=("${PIPESTATUS[@]}")

    if [[ ${#BP_PIPESTATUS[@]} -ge ${#_omp_pipestatus[@]} ]]; then
        _omp_pipestatus=("${BP_PIPESTATUS[@]}")
    fi

    _omp_stack_count=$((${#DIRSTACK[@]} - 1))

    _omp_execution_time=-1
    if [[ $_omp_start_time ]]; then
        local omp_now=$("$_omp_executable" get millis)
        _omp_execution_time=$((omp_now - _omp_start_time))
        _omp_no_status=false
    fi
    _omp_start_time=''

    if [[ ${_omp_pipestatus[-1]} != "$_omp_status" ]]; then
        _omp_pipestatus=("$_omp_status")
    fi

    set_poshcontext
    _omp_set_cursor_position

    PS1='$(_omp_get_primary)'
    PS2='$(_omp_get_secondary)'

    # Ensure that command substitution works in a prompt string.
    shopt -s promptvars

    return $_omp_status
}

function _omp_install_hook() {
    [[ $TERM = linux ]] && return

    local cmd
    for cmd in "${PROMPT_COMMAND[@]}"; do
        if [[ $cmd = _omp_hook ]]; then
            return
        fi
    done
    PROMPT_COMMAND=(_omp_hook "${PROMPT_COMMAND[@]}")
}

_omp_install_hook
