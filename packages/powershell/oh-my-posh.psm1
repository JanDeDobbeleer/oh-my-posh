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

function Get-PoshDownloadUrl {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Version
    )

    $executable = ""
    if ($IsMacOS) {
        $executable = "posh-darwin-amd64"
    }
    elseif ($IsLinux) {
        # this is rather hacky but there's no other way for the time being
        $arch = uname -m
        if ($arch -eq 'aarch64') {
            $executable = "posh-linux-arm64"
        }
        elseif ($arch -eq 'armv7l') {
            $executable = "posh-linux-arm"
        }
        else {
            $executable = "posh-linux-amd64"
        }
    }
    else {
        $arch = (Get-CimInstance -Class Win32_Processor -Property Architecture).Architecture
        switch ($arch) {
            0 { $executable = "posh-windows-386.exe" } # x86
            5 { $executable = "posh-windows-arm64.exe" } # ARM
            9 { $executable = "posh-windows-amd64.exe" } # x64
            12 { $executable = "posh-windows-amd64.exe" } # x64 emulated on Surface Pro X
        }
    }
    if ($executable -eq "") {
        throw "oh-my-posh: Unsupported architecture: $arch"
    }
    return "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$executable"
}

function Get-PoshExecutable {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Url,
        [Parameter(Mandatory = $true)]
        [string]
        $Destination
    )

    Invoke-WebRequest $Url -Out $Destination
    if (-Not (Test-Path $executable)) {
        # This should only happen with a corrupt installation
        throw "Executable at $executable was not found, please try importing oh-my-posh again."
    }
    # Set the right binary to executable before doing anything else
    # Permissions don't need to be set on Windows
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        return
    }
    chmod a+x $executable 2>&1
}

function Get-PoshCommand {
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        return Join-Path -Path $env:POSH_PATH -ChildPath "oh-my-posh.exe"
    }
    return Join-Path -Path $env:POSH_PATH -ChildPath "oh-my-posh"
}

function Sync-PoshThemes {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Version
    )

    Write-Host "Downloading oh-my-posh themes for $Version"
    $tmp = New-TemporaryFile | Rename-Item -NewName { $_ -replace 'tmp$', 'zip' } -PassThru
    $themesUrl = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/themes.zip"
    Invoke-WebRequest -OutFile $tmp $themesUrl
    $destination = $env:POSH_THEMES_PATH
    $tmp | Microsoft.PowerShell.Archive\Expand-Archive -DestinationPath $destination -Force
    $tmp | Remove-Item
}

function Sync-PoshArtifacts {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Version
    )

    $executable = Get-PoshCommand
    if (-not (Test-Path $executable)) {
        Write-Host "Downloading oh-my-posh executable for $Version"
        $url = Get-PoshDownloadUrl -Version $Version
        Get-PoshExecutable -Url $url -Destination $executable
        Sync-PoshThemes -Version $Version
        return
    }
    $poshVersion = & $executable --version
    if ($poshVersion -eq "development") {
        Write-Warning "omp development version detected"
        return
    }
    if ($poshVersion -eq $Version) {
        return
    }
    Write-Host "Updating oh-my-posh executable to $Version"
    $url = Get-PoshDownloadUrl -Version $Version
    Get-PoshExecutable -Url $url -Destination $executable
    Sync-PoshThemes -Version $Version
}

Set-PoshRootPath
$env:PATH = $env:POSH_PATH + [System.IO.Path]::PathSeparator + $env:PATH
$env:POSH_THEMES_PATH = Join-Path -Path $env:POSH_PATH -ChildPath "themes"
$moduleVersion = Split-Path -Leaf $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase
Sync-PoshArtifacts -Version $moduleVersion

# Legacy functions

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
        $config = "$($env:POSH_THEMES_PATH)/jandedobbeleer.omp.json"
    }

    # Workaround for get-location/push-location/pop-location from within a module
    # https://github.com/PowerShell/PowerShell/issues/12868
    # https://github.com/JanDeDobbeleer/oh-my-posh2/issues/113
    $global:omp_global_sessionstate = $PSCmdlet.SessionState

    $poshCommand = Get-PoshCommand
    (& $poshCommand init pwsh --config="$config") | Invoke-Expression
}
