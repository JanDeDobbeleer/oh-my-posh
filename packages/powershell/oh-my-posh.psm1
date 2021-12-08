$env:POSH_PATH = "$((Get-Item $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase).Parent.FullName)"
$env:POSH_THEMES_PATH = $env:POSH_PATH + "/themes"
$env:PATH = "$env:POSH_PATH;$env:PATH"

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
    $extension = ""
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        $extension = ".exe"
    }
    return "$((Get-Item $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase).Parent.FullName)/oh-my-posh$extension"
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
    $tmp | Expand-Archive -DestinationPath $destination -Force
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
    if ($poshVersion -eq $Version) {
        return
    }
    Write-Host "Updating oh-my-posh executable to $Version"
    $url = Get-PoshDownloadUrl -Version $Version
    Get-PoshExecutable -Url $url -Destination $executable
    Sync-PoshThemes -Version $Version
}

try {
    $moduleVersion = Split-Path -Leaf $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase
    Sync-PoshArtifacts -Version $moduleVersion
}
catch {
    $message = @'
Oh My Posh is unable to download and store the latest version.
In case you installed using AllUsers and are a non-admin user,
please run the following command as an administrator:

Import-Module oh-my-posh
'@
    Write-Error $message
    exit 1
}

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
    (& $poshCommand --init --shell=pwsh --config="$config") | Invoke-Expression
}
