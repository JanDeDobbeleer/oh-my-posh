export POSH_THEME=::CONFIG::
export POWERLINE_COMMAND="oh-my-posh"
export CONDA_PROMPT_MODIFIER=false
omp_start_time=""

# start timer on command start
PS0='${omp_start_time:0:$((omp_start_time="$(_omp_start_timer)",0))}'
# set secondary prompt
PS2="$(::OMP:: print secondary --config="$POSH_THEME" --shell=bash --shell-version="$BASH_VERSION")"

function _omp_start_timer() {
    ::OMP:: get millis
}

function _omp_hook() {
    local ret=$?
    local omp_stack_count=$((${#DIRSTACK[@]} - 1))
    local omp_elapsed=-1
    if [[ -n "$omp_start_time" ]]; then
        local omp_now=$(::OMP:: get millis)
        omp_elapsed=$((omp_now-omp_start_time))
        omp_start_time=""
    fi
    PS1="$(::OMP:: print primary --config="$POSH_THEME" --shell=bash --shell-version="$BASH_VERSION" --error="$ret" --execution-time="$omp_elapsed" --stack-count="$omp_stack_count" | tr -d '\0')"
    return $ret
}

if [ "$TERM" != "linux" ] && [ -x "$(command -v ::OMP::)" ] && ! [[ "$PROMPT_COMMAND" =~ "_omp_hook" ]]; then
    PROMPT_COMMAND="_omp_hook; $PROMPT_COMMAND"
fi
