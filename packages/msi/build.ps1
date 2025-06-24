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
    [switch]$Sign,
    [switch]$Copy
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$PSDefaultParameterValues['Out-File:Encoding'] = 'UTF8'

Write-Host "Building MSI for $Architecture with version $Version"

Write-Host "Setting up folders"

New-Item -Path "." -Name "dist" -ItemType Directory -ErrorAction SilentlyContinue
New-Item -Path "." -Name "out" -ItemType Directory -ErrorAction SilentlyContinue

if ($Copy) {
    switch ($Architecture) {
        'x86' { $file = "posh-windows-386.exe" }
        'x64' { $file = "posh-windows-amd64.exe" }
        Default { $file = "posh-windows-$Architecture.exe" }
    }

    Write-Host "Copying $file to ./dist/oh-my-posh.exe"

    # copy the correct architecture to ./dist
    Copy-Item -Path "../../dist/$file" -Destination "./dist/oh-my-posh.exe"
}

# variables
$env:VERSION = $Version

Write-Host "Creating MSI package"

# create MSI
$fileName = "install-$Architecture.msi"
$installer = "$PWD/out/$fileName" -replace '\\', '/'
wix build -arch $Architecture -out $installer .\oh-my-posh.wxs

if ($Sign) {
    Write-Host "Signing MSI"

    # setup dependencies
    nuget.exe install Microsoft.Trusted.Signing.Client -Version 1.0.92 -x
    $signtoolDlib = "$PWD/Microsoft.Trusted.Signing.Client/bin/x64/Azure.CodeSigning.Dlib.dll"
    $signtool = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/signtool.exe"

    # clean paths
    $signtool = $signtool -Replace '\\', '/'
    $signtoolDlib = $signtoolDlib -Replace '\\', '/'

    # sign installer
    & $signtool sign /v /debug /fd SHA256 /tr 'http://timestamp.acs.microsoft.com' /td SHA256 /dlib "$signtoolDlib" /dmdf ../../src/metadata.json "$installer"
}

Write-Host "Creating MSIX package"

# msix
$current = $PWD -replace '\\', '/'
$manifest = "$current/appxmanifest.xml"
$mappingTxt = "$current/mapping.txt"
$installerMSIX = "$current/out/$($filename)x"

[xml]$XmlDocument = Get-Content $manifest

$XmlDocument.Package.Identity.Version = "$Version.0"
$XmlDocument.Package.Identity.ProcessorArchitecture = $Architecture

$XmlDocument.Save($manifest)

Write-Host "Creating mapping file"

$mapping = @"
[ResourceMetadata]
"ResourceDimensions"                    "language-en-us"
"ResourceId"                            "English"

[Files]
"./dist/oh-my-posh.exe" "oh-my-posh.exe"
"./icons/icon.png" "/icons/icon.png"
"./icons/44.png" "/icons/44.png"
"@.Trim()

$stringBuilder = New-Object -TypeName System.Text.StringBuilder
$stringBuilder.Append($mapping) | Out-Null

# Add all theme files to the mapping and add to $mapping
Get-ChildItem -Path "../../themes" -Recurse -File -Filter "*.omp.*" | ForEach-Object {
    $stringBuilder.Append("`n`"../../themes/$($_.Name)`" `"/themes/$($_.Name)`"") | Out-Null
}

$stringBuilder.ToString() | Out-File -FilePath $mappingTxt

$makeappx = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/makeappx.exe"

& "$makeappx" pack /p $installerMSIX /v /o /m $manifest /f $mappingTxt

if ($Sign) {
    Write-Host "Signing MSIX"
    & "$signtool" sign /v /debug /fd SHA256 /tr 'http://timestamp.acs.microsoft.com' /td SHA256 /dlib "$signtoolDlib" /dmdf ../../src/metadata.json "$installerMSIX"
}

Write-Host "Creating hash files"

function Set-FileHash {
    param (
        [string]$File
    )
    $hash = Get-FileHash -Path $File -Algorithm SHA256
    $hash.Hash | Out-File -Encoding 'UTF8' "$File.sha256"
}

Set-FileHash -File $installer
Set-FileHash -File $installerMSIX

Write-Host "Finished building MSI and MSIX packages"
