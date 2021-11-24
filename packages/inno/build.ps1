Param
(
    [parameter(Mandatory = $true)]
    [ValidateSet('amd64', 'arm64')]
    [System.String]$Architecture,
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

New-Item -Path "." -Name "bin" -ItemType Directory
Copy-Item -Path "../../themes" -Destination "./bin" -Recurse

# download the files and pack them
@{file = "posh-windows-$Architecture.exe"; name = "oh-my-posh.exe" } | ForEach-Object -Process {
    $download = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$($_.file)"
    Invoke-WebRequest $download -Out "./bin/$($_.name)"
}
# license
Invoke-WebRequest "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v$Version/COPYING" -Out "./bin/COPYING.txt"
$content = Get-Content '.\oh-my-posh.iss' -Raw
$content = $content.Replace('<VERSION>', $Version)
$ISSName = ".oh-my-posh-$Architecture-$Version.iss"
$content | Out-File -Encoding 'UTF8' $ISSName
# package content
$installer = "install-$Architecture"
ISCC.exe /F$installer $ISSName
# get hash
$zipHash = Get-FileHash "Output/$installer.exe" -Algorithm SHA256
$zipHash.Hash | Out-File -Encoding 'UTF8' "Output/$installer.exe.sha256"
