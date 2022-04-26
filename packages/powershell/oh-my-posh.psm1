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

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )
}
