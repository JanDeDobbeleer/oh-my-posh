<#
.SYNOPSIS
    Builds MSI and MSIX packages for Oh My Posh.

.DESCRIPTION
    This script creates MSI and MSIX installer packages for Oh My Posh with the specified architecture and version.
    It can optionally copy the executable, sign the packages, and generate hash files for verification.

.PARAMETER Architecture
    The target architecture for the package. Must be either 'x64' or 'arm64'.

.PARAMETER Version
    The version number to assign to the package (e.g., "1.2.3").

.PARAMETER SDKVersion
    The Windows SDK version to use for signing and packaging tools. Defaults to "10.0.26100.0".

.PARAMETER Sign
    When specified, signs the MSI and MSIX packages using Azure Code Signing.

.PARAMETER Copy
    When specified, copies the appropriate executable from the dist folder before packaging.

.EXAMPLE
    .\build.ps1 -Architecture x64 -Version "1.2.3" -Copy

    Creates MSI and MSIX packages for x64 architecture with version 1.2.3, copying the executable first.

.EXAMPLE
    .\build.ps1 -Architecture arm64 -Version "1.2.3" -Sign -Copy

    Creates and signs MSI and MSIX packages for arm64 architecture with version 1.2.3.

.OUTPUTS
    Creates the following files in the 'out' directory:
    - install-{Architecture}.msi
    - install-{Architecture}.msix
    - Hash files (.sha256) for verification

.NOTES
    Requires WiX toolset for MSI creation and Windows SDK for MSIX packaging and signing.
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('x64', 'arm64')]
    [string]$Architecture,

    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [string]$Version,

    [Parameter()]
    [ValidateNotNullOrEmpty()]
    [string]$SDKVersion = "10.0.26100.0",

    [Parameter()]
    [switch]$Sign,

    [Parameter()]
    [switch]$Copy
)

# Set error handling preferences
$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true
$PSDefaultParameterValues['Out-File:Encoding'] = 'UTF8'

#region Helper Functions

function Initialize-SigningEnvironment {
    <#
    .SYNOPSIS
        Sets up the signing environment and returns signing tool paths.

    .PARAMETER SDKVersion
        The Windows SDK version to use.

    .OUTPUTS
        Hashtable containing signtool and signtoolDlib paths.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$SDKVersion
    )

    try {
        Write-Verbose "Setting up signing environment" -Verbose

        # Install Microsoft.Trusted.Signing.Client
        nuget.exe install Microsoft.Trusted.Signing.Client -Version 1.0.92 -x

        $signtoolDlib = "$PWD/Microsoft.Trusted.Signing.Client/bin/x64/Azure.CodeSigning.Dlib.dll" -replace '\\', '/'
        $signtool = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/signtool.exe" -replace '\\', '/'

        # Validate tools exist
        if (-not (Test-Path $signtool)) {
            throw "signtool.exe not found at: $signtool"
        }
        if (-not (Test-Path $signtoolDlib)) {
            throw "Azure.CodeSigning.Dlib.dll not found at: $signtoolDlib"
        }

        return @{
            SignTool = $signtool
            SignToolDlib = $signtoolDlib
        }
    }
    catch {
        Write-Error "Failed to initialize signing environment: $_"
        throw
    }
}

function Invoke-PackageSigning {
    <#
    .SYNOPSIS
        Signs a package using Azure Code Signing.

    .PARAMETER PackagePath
        The path to the package to sign.

    .PARAMETER SigningTools
        Hashtable containing signing tool paths from Initialize-SigningEnvironment.
    #>
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateScript({Test-Path $_})]
        [string]$PackagePath,

        [Parameter(Mandatory = $true)]
        [hashtable]$SigningTools
    )

    try {
        $packageName = Split-Path $PackagePath -Leaf
        Write-Verbose "Signing package: $packageName" -Verbose

        & $SigningTools.SignTool sign /v /debug /d "Oh My Posh" /fd SHA256 /tr 'http://timestamp.acs.microsoft.com' /td SHA256 /dlib $SigningTools.SignToolDlib /dmdf ../../src/metadata.json $PackagePath

        Write-Verbose "Successfully signed: $packageName" -Verbose
    }
    catch {
        Write-Error "Failed to sign package ${PackagePath}: ${_}"
        throw
    }
}

#endregion

#region Main Script

Write-Verbose "Building MSI for $Architecture with version $Version" -Verbose
Write-Verbose "Setting up output directories" -Verbose

try {
    New-Item -Path "." -Name "dist" -ItemType Directory -ErrorAction SilentlyContinue | Out-Null
    New-Item -Path "." -Name "out" -ItemType Directory -ErrorAction SilentlyContinue | Out-Null
}
catch {
    Write-Error "Failed to create output directories: ${_}"
    throw
}

if ($Copy) {
    $sourceFile = switch ($Architecture) {
        'x64' { "posh-windows-amd64.exe" }
        Default { "posh-windows-$Architecture.exe" }
    }

    Write-Verbose "Copying $sourceFile to ./dist/oh-my-posh.exe" -Verbose

    try {
        $sourcePath = "../../dist/$sourceFile"
        if (-not (Test-Path $sourcePath)) {
            throw "Source file not found: $sourcePath"
        }
        Copy-Item -Path $sourcePath -Destination "./dist/oh-my-posh.exe" -Force
    }
    catch {
        Write-Error "Failed to copy executable: $_"
        throw
    }
}

# Set version environment variable for WiX
$env:VERSION = $Version

Write-Verbose "Creating MSI package" -Verbose

try {
    # Define MSI package paths
    $msiFileName = "install-$Architecture.msi"
    $msiPackagePath = "$PWD/out/$msiFileName" -replace '\\', '/'

    Write-Verbose "Building MSI: $msiPackagePath" -Verbose
    wix build -arch $Architecture -out $msiPackagePath .\oh-my-posh.wxs

    if (-not (Test-Path $msiPackagePath)) {
        throw "MSI package was not created successfully"
    }
}
catch {
    Write-Error "Failed to create MSI package: ${_}"
    throw
}

if ($Sign) {
    $signingTools = Initialize-SigningEnvironment -SDKVersion $SDKVersion
    Invoke-PackageSigning -PackagePath $msiPackagePath -SigningTools $signingTools
}

Write-Verbose "Creating MSIX package" -Verbose

try {
    # Define MSIX package paths and files
    $currentPath = $PWD -replace '\\', '/'
    $manifestPath = "$currentPath/appxmanifest.xml"
    $mappingFilePath = "$currentPath/mapping.txt"
    $msixPackagePath = "$currentPath/out/$($msiFileName)x"
    $makeappxPath = "C:/Program Files (x86)/Windows Kits/10/bin/$SDKVersion/x64/makeappx.exe"

    # Validate required files exist
    if (-not (Test-Path $manifestPath)) {
        throw "Manifest file not found: $manifestPath"
    }
    if (-not (Test-Path $mappingFilePath)) {
        throw "Mapping file not found: $mappingFilePath"
    }
    if (-not (Test-Path $makeappxPath)) {
        throw "makeappx.exe not found at: $makeappxPath"
    }

    # Update manifest with version and architecture
    [xml]$manifestDocument = Get-Content $manifestPath
    $manifestDocument.Package.Identity.Version = "$Version.0"
    $manifestDocument.Package.Identity.ProcessorArchitecture = $Architecture
    $manifestDocument.Save($manifestPath)

    # Build MSIX package
    Write-Verbose "Building MSIX: $msixPackagePath" -Verbose
    & "$makeappxPath" pack /p $msixPackagePath /v /o /m $manifestPath /f $mappingFilePath

    if (-not (Test-Path $msixPackagePath)) {
        throw "MSIX package was not created successfully"
    }
}
catch {
    Write-Error "Failed to create MSIX package: ${_}"
    throw
}

if ($Sign) {
    if (-not $signingTools) {
        $signingTools = Initialize-SigningEnvironment -SDKVersion $SDKVersion
    }
    Invoke-PackageSigning -PackagePath $msixPackagePath -SigningTools $signingTools
}

Write-Verbose "Successfully completed building MSI and MSIX packages" -Verbose

#endregion
