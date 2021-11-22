# Activate oh-my-posh prompt:
Import-Module oh-my-posh
Set-PoshPrompt -Theme ${env:POSHTHEMES_ROOT}/${env:DEFAULT_POSH_THEME}.omp.json

# NOTE: You can override the above env vars from the devcontainer.json "args" under the "build" key.
