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

$moduleDir = "./oh-my-posh"
New-Item -Path $moduleDir -ItemType Directory
# set the actual version number
(Get-Content './oh-my-posh.psd1' -Raw).Replace('0.0.0.1', $ModuleVersion) | Out-File -Encoding 'UTF8' "$moduleDir/oh-my-posh.psd1"
Copy-Item "./oh-my-posh.psm1" -Destination $moduleDir
Push-Location -Path $moduleDir
# publish the module
if ($RepositoryAPIKey) {
    Publish-Module -Path . -Repository $Repository -NuGetApiKey $RepositoryAPIKey -Verbose
} else {
    Publish-Module -Path . -Repository $Repository -Verbose
}
Pop-Location
Remove-Item -Path $moduleDir -Force -Recurse
