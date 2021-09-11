Param
(
    [parameter(Mandatory=$true)]
    [string]
    $BinVersion,
    [parameter(Mandatory=$true)]
    [string]
    $ModuleVersion,
    [parameter(Mandatory=$true)]
    [string]
    $Repository,
    [parameter(Mandatory=$false)]
    [string]
    $RepositoryAPIKey
)

# set the actual version number
(Get-Content '.\oh-my-posh.psd1' -Raw).Replace('0.0.0.1', $ModuleVersion) | Out-File -Encoding 'UTF8' '.\oh-my-posh.psd1'
# copy all themes into the module folder
Copy-Item -Path "../../../themes" -Destination "./themes" -Recurse
# fetch all the binaries from the version's GitHub release
New-Item -Path "./" -Name "bin" -ItemType "directory"
"posh-windows-amd64.exe", "posh-windows-386.exe", "posh-windows-arm64.exe", "posh-darwin-amd64", "posh-linux-amd64", "posh-linux-arm" | ForEach-Object -Process {
    $download = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$BinVersion/$_"
    Invoke-WebRequest $download -Out "./bin/$_"
}
# publish the module
if ($RepositoryAPIKey) {
    Publish-Module -Path . -Repository $Repository -NuGetApiKey $RepositoryAPIKey -Verbose
} else {
    Publish-Module -Path . -Repository $Repository -Verbose
}
# reset module version (for local testing only as we don't want PR's with changed version numbers all the time)
(Get-Content '.\oh-my-posh.psd1' -Raw).Replace($ModuleVersion, '0.0.0.1') | Out-File -Encoding 'UTF8' '.\oh-my-posh.psd1'
Remove-Item "./bin" -Recurse -Force
Remove-Item "./themes" -Recurse -Force

