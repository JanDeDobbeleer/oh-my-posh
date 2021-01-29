
Param
(
    [parameter(Mandatory = $true)]
    [string]
    $Version,
    [parameter(Mandatory = $true)]
    [string]
    $Hash,
    [parameter(Mandatory = $false)]
    [string]
    $Token
)

function Set-Version {
    param (
        [parameter(Mandatory = $true)]
        [string]
        $FileName,
        [parameter(Mandatory = $true)]
        [string]
        $Version,
        [parameter(Mandatory = $true)]
        [string]
        $Hash
    )
    $content = Get-Content $FileName -Raw
    $content = $content.Replace('<VERSION>', $Version)
    $content = $content.Replace('<HASH>', $Hash)
    $content | Out-File -Encoding 'UTF8' "./$Version/$FileName"
}

New-Item -Path $PWD -Name $Version -ItemType "directory"
# Get all files inside the folder and adjust the version/hash
Get-ChildItem '*.yaml' | ForEach-Object -Process {
    Set-Version -FileName $_.Name -Version $Version -Hash $hash
}
if (-not $Token) {
    return
}
# Get the latest wingetcreate exe
# Replace with the following once https://github.com/microsoft/winget-create/issues/38 is resolved:
# Invoke-WebRequest https://aka.ms/wingetcreate/latest -OutFile wingetcreate.exe
Invoke-WebRequest 'https://github.com/JanDeDobbeleer/winget-create/releases/latest/download/wingetcreate.zip' -OutFile wingetcreate.zip
Expand-Archive -LiteralPath wingetcreate.zip -DestinationPath wingetcreate
$wingetcreate = Resolve-Path -Path wingetcreate
$env:Path += ";$($wingetcreate.Path)"
# Create the PR
WingetCreateCLI.exe submit --token $Token $Version
