Param
(
    [parameter(Mandatory = $true)]
    [ValidateSet('x64', 'arm64', 'x86')]
    [System.String]$Architecture,
    [parameter(Mandatory = $true)]
    [string]
    $Version,
    [parameter(Mandatory = $false)]
    [string]
    $SDKVersion = "10.0.22621.0",
    [switch]$Download,
    [switch]$Sign
)

$PSDefaultParameterValues['Out-File:Encoding'] = 'UTF8'

New-Item -Path "." -Name "dist" -ItemType Directory -ErrorAction SilentlyContinue
New-Item -Path "." -Name "out" -ItemType Directory -ErrorAction SilentlyContinue

if ($Download) {
    # download the executable
    $file = "posh-windows-$Architecture.exe"
    $name = "oh-my-posh.exe"
    $url = "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$($file)"
    Invoke-WebRequest $url -Out "./dist/$($name)"
}

# variables
$env:VERSION = $Version

# create MSI
$installer = "./out/install-$Architecture.msi"
wix build -arch $Architecture -out $installer .\oh-my-posh.wxs

if ($Sign) {
    # setup dependencies
    nuget.exe install Microsoft.Trusted.Signing.Client -Version 1.0.60 -x
    $signtoolDlib = "$PWD/Microsoft.Trusted.Signing.Client/bin/x64/Azure.CodeSigning.Dlib.dll"
    $signtool = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/signtool.exe"

    # clean paths
    $signtool = $signtool -Replace '\\', '/'
    $signtoolDlib = $signtoolDlib -Replace '\\', '/'

    # sign installer
    & "$signtool" sign /v /debug /fd SHA256 /tr 'http://timestamp.acs.microsoft.com' /td SHA256 /dlib "$signtoolDlib" /dmdf ../../src/metadata.json $installer
}

# get hash
$zipHash = Get-FileHash $installer -Algorithm SHA256
$zipHash.Hash | Out-File -Encoding 'UTF8' "$installer.sha256"
