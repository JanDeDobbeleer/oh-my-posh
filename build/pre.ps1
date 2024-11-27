Param
(
    [string]
    $Version,
    [parameter(Mandatory = $false)]
    [string]
    $SDKVersion = "10.0.22621.0"
)

git config --global user.name "GitHub Actions"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
git tag $Version --force

$PSDefaultParameterValues['Out-File:Encoding'] = 'UTF8'

$shaSigningKeyLocation = Join-Path -Path $env:RUNNER_TEMP -ChildPath sha_signing_key.pem
$env:SIGNING_KEY > $shaSigningKeyLocation
Write-Output "SHA_SIGNING_KEY_LOCATION=$shaSigningKeyLocation" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append

# install code signing dlib
nuget.exe install Microsoft.Trusted.Signing.Client -Version 1.0.60 -ExcludeVersion -OutputDirectory $env:RUNNER_TEMP
Write-Output "SIGNTOOLDLIB=$env:RUNNER_TEMP/Microsoft.Trusted.Signing.Client/bin/x64/Azure.CodeSigning.Dlib.dll" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append

# requires Windows Dev Kit 10.0.22621.0
$signtool = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/signtool.exe"
Write-Output "SIGNTOOL=$signtool" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append

# openssl
$openssl = 'C:/Program Files/Git/usr/bin/openssl.exe'
Write-Output "OPENSSL=$openssl" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append
