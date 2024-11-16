Param
(
    [parameter(Mandatory = $true)]
    [ValidateSet('amd64', 'arm64', '386')]
    [System.String]$Architecture,
    [parameter(Mandatory = $true)]
    [string]
    $Version
)

$PSDefaultParameterValues['Out-File:Encoding']='UTF8'

# setup dependencies
nuget.exe install Microsoft.Trusted.Signing.Client -Version 1.0.60 -x
$signtoolDlib = "$PWD/Microsoft.Trusted.Signing.Client/bin/x64/Azure.CodeSigning.Dlib.dll"
$signtool = 'C:/Program Files (x86)/Windows Kits/10/bin/10.0.22621.0/x64/signtool.exe'

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

# clean paths
$signtool = $signtool -Replace '\\', '/'
$signtoolDlib = $signtoolDlib -Replace '\\', '/'

# package content
$installer = "install-$Architecture"
ISCC.exe /F$installer $ISSName

# sign installer
& "$signtool" sign /v /debug /fd SHA256 /tr 'http://timestamp.acs.microsoft.com' /td SHA256 /dlib "$signtoolDlib" /dmdf ../../src/metadata.json "./Output/$installer.exe"

# get hash
$zipHash = Get-FileHash "Output/$installer.exe" -Algorithm SHA256
$zipHash.Hash | Out-File -Encoding 'UTF8' "Output/$installer.exe.sha256"
