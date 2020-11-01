Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version,
    [parameter(Mandatory = $true)]
    [string]
    $LinuxSHA,
    [parameter(Mandatory = $true)]
    [string]
    $WindowsSHA,
    [parameter(Mandatory = $true)]
    [string]
    $ThemesSHA
)

$content = Get-Content '.\oh-my-posh.json' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content = $content.Replace('<HASH_LINUX>', $LinuxSHA)
$content = $content.Replace('<HASH_WINDOWS>', $WindowsSHA)
$fileHash = Get-FileHash post-install.ps1 -Algorithm SHA256
$content = $content.Replace('<HASH_INSTALL_SCRIPT>', $fileHash.Hash)
$content = $content.Replace('<HASH_THEMES>', $ThemesSHA)
$content | Out-File -Encoding 'UTF8' '.\oh-my-posh.json'
