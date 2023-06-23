
Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version,
    [parameter(Mandatory = $false)]
    [string]
    $Token
)

function Get-HashForArchitecture {
    param (
        [parameter(Mandatory = $true)]
        [string]
        $Architecture,
        [parameter(Mandatory = $true)]
        [string]
        $Version
    )
    $hash = (new-object Net.WebClient).DownloadString("https://github.com/JanDeDobbeleer/oh-my-posh/releases/download/v$Version/install-$Architecture.exe.sha256")
    return $hash.Trim()
}

function Write-MetaData {
    param (
        [parameter(Mandatory = $true)]
        [string]
        $FileName,
        [parameter(Mandatory = $true)]
        [string]
        $Version,
        [parameter(Mandatory = $true)]
        [string]
        $HashAmd64,
        [parameter(Mandatory = $true)]
        [string]
        $HashArm64,
        [parameter(Mandatory = $true)]
        [string]
        $Hash386
    )
    $content = Get-Content $FileName -Raw
    $content = $content.Replace('<VERSION>', $Version)
    $content = $content.Replace('<HASH-AMD64>', $HashAmd64)
    $content = $content.Replace('<HASH-ARM64>', $HashArm64)
    $content = $content.Replace('<HASH-386>', $Hash386)
    $date = Get-Date -Format "yyyy-MM-dd"
    $content = $content.Replace('<DATE>', $date)
    $content | Out-File -Encoding 'UTF8' "./$Version/$FileName"
}

New-Item -Path $PWD -Name $Version -ItemType "directory"
# Get all files inside the folder and adjust the version/hash
$HashAmd64 = Get-HashForArchitecture -Architecture 'amd64' -Version $Version
$HashArm64 = Get-HashForArchitecture -Architecture 'arm64' -Version $Version
$Hash386 = Get-HashForArchitecture -Architecture '386' -Version $Version
Get-ChildItem '*.yaml' | ForEach-Object -Process {
    Write-MetaData -FileName $_.Name -Version $Version -HashAmd64 $HashAmd64 -HashArm64 $HashArm64 -Hash386 $Hash386
}
if (-not $Token) {
    return
}
# Install the latest wingetcreate exe
# Need to do things this way, see https://github.com/PowerShell/PowerShell/issues/13138
Import-Module Appx -UseWindowsPowerShell

# Download and install C++ Runtime framework package.
$vcLibsBundleFile = "$env:TEMP\Microsoft.VCLibs.Desktop.appx"
Invoke-WebRequest https://aka.ms/Microsoft.VCLibs.x64.14.00.Desktop.appx -OutFile $vcLibsBundleFile
Add-AppxPackage $vcLibsBundleFile

# Download Winget-Create msixbundle, install, and execute update.
$appxBundleFile = "$env:TEMP\wingetcreate.msixbundle"
Invoke-WebRequest https://aka.ms/wingetcreate/latest/msixbundle -OutFile $appxBundleFile
Add-AppxPackage $appxBundleFile

# Create the PR
wingetcreate submit --token $Token $Version
