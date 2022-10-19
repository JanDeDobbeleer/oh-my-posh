Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version
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

New-Item -Path "." -Name "dist" -ItemType "directory"

$HashAmd64 = Get-HashForArchitecture -Architecture 'amd64' -Version $Version
$Hash386 = Get-HashForArchitecture -Architecture '386' -Version $Version

$content = Get-Content '.\oh-my-posh.json' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content = $content.Replace('<HASH-AMD64>', $HashAmd64)
$content = $content.Replace('<HASH-386>', $Hash386)
$content | Out-File -Encoding 'UTF8' './dist/oh-my-posh.json'
