Import-Module posh-git
Import-Module PSFzf -ArgumentList 'Ctrl+t', 'Ctrl+r'
Import-Module z
Import-Module Terminal-Icons

Set-PSReadlineKeyHandler -Key Tab -Function MenuComplete

$env:POSH_GIT_ENABLED=$true
oh-my-posh init pwsh | Invoke-Expression
