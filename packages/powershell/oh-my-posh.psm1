Write-Host @'
Hey friend

In an effort to grow oh-my-posh, the decision was made to no
longer support the PowerShell module. Over the past year, the
added benefit of the module disappeared, while the burden of
maintaining it increased.

However, this doesn't mean oh-my-posh disappears from your
terminal, it just means that you'll have to use a different
tool to install it.

All you need to do, is follow the migration guide here:

https://ohmyposh.dev/docs/migrating
'@

function Get-CachePath {
    [CmdletBinding()]
    param (
        [Parameter()]
        [string]
        $Path
    )

    if (-Not (Test-Path -Path $Path)) {
        return ""
    }
    $child = "oh-my-posh"
    $cachePath = Join-Path -Path $Path -ChildPath $child
    if (Test-Path -Path $cachePath) {
        return $cachePath
    }
    $null = New-Item -Path $Path -Name $child -ItemType Directory
    return $cachePath
}

function Set-PoshRootPath {
    if ($env:POSH_PATH) {
        return
    }
    $path = Get-CachePath -Path $env:LOCALAPPDATA
    if ($path) {
        $env:POSH_PATH = $path
        return
    }
    $path = Get-CachePath -Path $env:XDG_CACHE_HOME
    if ($path) {
        $env:POSH_PATH = $path
        return
    }
    $homeCache = Join-Path -Path $env:HOME -ChildPath ".cache"
    $path = Get-CachePath -Path $homeCache
    if ($path) {
        $env:POSH_PATH = $path
        return
    }
    $env:POSH_PATH = $env:HOME
}

Set-PoshRootPath
$env:PATH = $env:POSH_PATH + [System.IO.Path]::PathSeparator + $env:PATH
$env:POSH_THEMES_PATH = Join-Path -Path $env:POSH_PATH -ChildPath "themes"

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )

    $config = ""
    if (Test-Path "$($env:POSH_THEMES_PATH)/$Theme.omp.json") {
        $path = "$($env:POSH_THEMES_PATH)/$Theme.omp.json"
        $config = (Resolve-Path -Path $path).ProviderPath
    }
    elseif (Test-Path $Theme) {
        $config = (Resolve-Path -Path $Theme).ProviderPath
    }
    else {
        $config = "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v7.71.2/themes/default.omp.json"
    }

    # Workaround for get-location/push-location/pop-location from within a module
    # https://github.com/PowerShell/PowerShell/issues/12868
    # https://github.com/JanDeDobbeleer/oh-my-posh2/issues/113
    $global:OMP_GLOBAL_SESSIONSTATE = $PSCmdlet.SessionState

    oh-my-posh init pwsh --config="$config" | Invoke-Expression
}
