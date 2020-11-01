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
    $WindowsSHA
)

$content = Get-Content '.\scoop.json' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content = $content.Replace('<HASH_LINUX>', $LinuxSHA)
$content = $content.Replace('<HASH_WINDOWS>', $WindowsSHA)
$fileHash = Get-FileHash post-install.ps1 -Algorithm SHA256
$content = $content.Replace('<HASH_INSTALL_SCRIPT>', $fileHash.Hash)
$content | Out-File -Encoding 'UTF8' '.\scoop.json'
