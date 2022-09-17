Param
(
    [parameter(Mandatory = $true)]
    [ValidateSet('amd64', 'arm64', '386')]
    [System.String]$Architecture,
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

# Get signing certificate
$pfxPath = Join-Path -Path $env:RUNNER_TEMP -ChildPath "cert.pfx"
$signtool = 'C:/Program Files (x86)/Windows Kits/10/bin/10.0.22000.0/x86/signtool.exe'
# create a base64 encoded value of your certificate using
# [convert]::ToBase64String((Get-Content -path "certificate.pfx" -AsByteStream))
# requires Windows Dev Kit 10.0.22000.0
$encodedBytes = [System.Convert]::FromBase64String($env:CERTIFICATE)
Set-Content -Path $pfxPath -Value $encodedBytes -AsByteStream

New-Item -Path "." -Name "bin" -ItemType Directory
Copy-Item -Path "../../themes" -Destination "./bin" -Recurse

# download the executable
$file = "posh-windows-$Architecture.exe"
$name = "oh-my-posh.exe"
$download = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$($file)"
Invoke-WebRequest $download -Out "./bin/$($name)"

# license
Invoke-WebRequest "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v$Version/COPYING" -Out "./bin/COPYING.txt"
$content = Get-Content '.\oh-my-posh.iss' -Raw
$content = $content.Replace('<VERSION>', $Version)
$ISSName = ".oh-my-posh-$Architecture-$Version.iss"
$content | Out-File -Encoding 'UTF8' $ISSName

# package content
$installer = "install-$Architecture"
ISCC.exe /F$installer "/Ssigntool=$signtool sign /f $pfxPath /p $env:CERTIFICATE_PASSWORD /fd SHA256 /t http://timestamp.digicert.com `$f" $ISSName
# get hash
$zipHash = Get-FileHash "Output/$installer.exe" -Algorithm SHA256
$zipHash.Hash | Out-File -Encoding 'UTF8' "Output/$installer.exe.sha256"
