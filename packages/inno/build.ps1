Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

New-Item -Path "." -Name "bin" -ItemType Directory
Copy-Item -Path "../../themes" -Destination "./bin" -Recurse

# download the files and pack them
@{name = 'posh-windows-amd64.exe' }, @{name = 'posh-linux-amd64' }, @{name = 'posh-windows-386.exe' } | ForEach-Object -Process {
    $download = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$($_.name)"
    Invoke-WebRequest $download -Out "./bin/$($_.name)"
}
# lisence
Invoke-WebRequest "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v$Version/COPYING" -Out "./bin/COPYING.txt"
$content = Get-Content '.\oh-my-posh.iss' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content | Out-File -Encoding 'UTF8' ".oh-my-posh-$Version.iss"
# package content
ISCC.exe ".oh-my-posh-$Version.iss"
# get hash
$zipHash = Get-FileHash 'Output/install.exe' -Algorithm SHA256
$zipHash.Hash | Out-File -Encoding 'UTF8' 'Output/install.exe.sha256'
