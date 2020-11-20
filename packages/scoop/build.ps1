Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

New-Item -Path "." -Name "bin" -ItemType Directory
Copy-Item -Path "../../themes" -Destination "./bin" -Recurse

# Download the files and pack them
@{name = 'posh-windows-amd64.exe'; outName = 'oh-my-posh.exe' }, @{name = 'posh-linux-amd64'; outName = 'oh-my-posh-wsl' } | ForEach-Object -Process {
    $download = "https://github.com/jandedobbeleer/oh-my-posh3/releases/download/v$Version/$($_.name)"
    Invoke-WebRequest $download -Out "./bin/$($_.outName)"
}
$compress = @{
    Path             = "./bin/*"
    CompressionLevel = "Fastest"
    DestinationPath  = "./posh-windows-wsl-amd64.7z"
}
Compress-Archive @compress
$zipHash = Get-FileHash ./posh-windows-wsl-amd64.7z -Algorithm SHA256
$content = Get-Content '.\oh-my-posh.json' -Raw
$content = $content.Replace('<VERSION>', $Version)
$content = $content.Replace('<HASH>', $zipHash.Hash)
$content | Out-File -Encoding 'UTF8' './oh-my-posh.json'

Remove-Item ./bin/ -Recurse
