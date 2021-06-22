
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
Invoke-WebRequest 'https://aka.ms/wingetcreate/latest/self-contained' -OutFile wingetcreate.exe
# Create the PR
./wingetcreate.exe submit --token $Token $Version
