Param
(
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
# publish the module
if ($RepositoryAPIKey) {
    Publish-Module -Path . -Repository $Repository -NuGetApiKey $RepositoryAPIKey -Verbose
} else {
    Publish-Module -Path . -Repository $Repository -Verbose
}
# reset module version (for local testing only as we don't want PR's with changed version numbers all the time)
(Get-Content '.\oh-my-posh.psd1' -Raw).Replace($ModuleVersion, '0.0.0.1') | Out-File -Encoding 'UTF8' '.\oh-my-posh.psd1'
Remove-Item "./themes" -Recurse -Force
