Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

New-Item -Path "." -Name "package/bin" -ItemType Directory
Copy-Item -Path "../../themes" -Destination "./package" -Recurse

# Download the files and pack them
@{name = 'posh-windows-amd64.exe'; outName = 'oh-my-posh.exe' } | ForEach-Object -Process {
    $download = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$($_.name)"
    Invoke-WebRequest $download -Out "./package/bin/$($_.outName)"
}
$compress = @{
    Path             = "./package/*"
    CompressionLevel = "Fastest"
    DestinationPath  = "./posh-windows-amd64.7z"
}
Compress-Archive @compress
$zipHash = Get-FileHash ./posh-windows-amd64.7z -Algorithm SHA256
$content = Get-Content '.\oh-my-posh.json' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content = $content.Replace('<HASH>', $zipHash.Hash)
$content | Out-File -Encoding 'UTF8' './oh-my-posh.json'
$zipHash.Hash | Out-File -Encoding 'UTF8' 'posh-windows-amd64.7z.sha256'

Remove-Item ./package/ -Recurse
