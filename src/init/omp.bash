export POSH_THEME=::CONFIG::

function _update_ps1() {
    PS1="$(::OMP:: --config $POSH_THEME --error $?)"
}

if [ "$TERM" != "linux" ] && [ -x "$(command -v ::OMP::)" ]; then
    PROMPT_COMMAND="_update_ps1; $PROMPT_COMMAND"
fi
