
<#
.SYNOPSIS
    Builds and updates WinGet package manifests for Oh My Posh.

.DESCRIPTION
    This script generates WinGet package manifests for a new version of Oh My Posh by:
    - Downloading installer metadata (hashes and signatures) from the CDN
    - Creating version-specific YAML manifests with updated hashes and version info
    - Optionally submitting a pull request to the WinGet repository using wingetcreate

.PARAMETER Version
    The version of Oh My Posh to build the package for. The 'v' prefix will be automatically removed if present.

.PARAMETER Token
    Optional GitHub token for submitting a pull request to the WinGet repository. If not provided, only the manifest files will be generated.

.EXAMPLE
    .\build.ps1 -Version "v16.0.0"
    Generates manifest files for version 16.0.0 without submitting a PR.

.EXAMPLE
    .\build.ps1 -Version "16.0.0" -Token "ghp_xxxxxxxxxxxx"
    Generates manifest files and submits a PR to the WinGet repository.
#>

[CmdletBinding()]
Param(
    [Parameter(Mandatory = $true, HelpMessage = "The version of Oh My Posh to build the package for")]
    [ValidateNotNullOrEmpty()]
    [string]$Version,

    [Parameter(Mandatory = $false, HelpMessage = "GitHub token for submitting WinGet PR")]
    [string]$Token
)

#region Helper Functions

<#
.SYNOPSIS
    Downloads installer metadata (hash or signature) for a specific architecture.

.DESCRIPTION
    Retrieves the SHA256 hash or signature data for Oh My Posh installer files from the official CDN.

.PARAMETER Architecture
    The architecture identifier (e.g., 'x64.msi.sha256', 'arm64.msix.sig.sha256').

.PARAMETER Version
    The version of Oh My Posh to get metadata for.

.OUTPUTS
    String containing the hash or signature data.
#>
function Get-InstallerMetadata {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Architecture,

        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Version
    )

    try {
        $uri = "https://cdn.ohmyposh.dev/releases/v$Version/install-$Architecture"
        Write-Verbose "Downloading metadata from: $uri"

        $metadata = Invoke-RestMethod -Uri $uri -ErrorAction Stop
        return $metadata.Trim()
    }
    catch {
        throw "Failed to download installer metadata for $Architecture`: $($_.Exception.Message)"
    }
}

<#
.SYNOPSIS
    Updates a YAML manifest template with version and hash information.

.DESCRIPTION
    Replaces placeholder values in YAML template files with actual version, hash, and date information.

.PARAMETER FileName
    Name of the YAML template file to process.

.PARAMETER Metadata
    Hashtable containing all the metadata values for replacement.
#>
function Update-ManifestFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$FileName,

        [Parameter(Mandatory = $true)]
        [hashtable]$Metadata
    )

    try {
        Write-Verbose "Processing manifest file: $FileName"

        # Read the template file
        $content = Get-Content -Path $FileName -Raw -ErrorAction Stop

        # Apply all replacements
        foreach ($placeholder in $Metadata.Keys) {
            if ($placeholder.StartsWith('<') -and $placeholder.EndsWith('>')) {
                $content = $content.Replace($placeholder, $Metadata[$placeholder])
            }
        }

        # Write the updated content to the version-specific directory
        $outputPath = Join-Path -Path $Metadata.Version -ChildPath $FileName
        $content | Out-File -FilePath $outputPath -Encoding UTF8 -ErrorAction Stop

        Write-Verbose "Successfully updated manifest: $outputPath"
    }
    catch {
        throw "Failed to update manifest file $FileName`: $($_.Exception.Message)"
    }
}

#endregion

#region Main Script Logic

try {
    # Clean version string (remove 'v' prefix if present)
    $Version = $Version.TrimStart('v')
    Write-Verbose "Processing version: $Version"

    # Create version directory
    $versionPath = Join-Path -Path $PWD -ChildPath $Version
    if (-not (Test-Path -Path $versionPath)) {
        New-Item -Path $versionPath -ItemType Directory -Force | Out-Null
        Write-Verbose "Created version directory: $versionPath"
    }

    # Define metadata files to download with their template placeholders
    $metadataFiles = @{
        '<HASH-AMD64>' = 'x64.msi.sha256'
        '<HASH-ARM64>' = 'arm64.msi.sha256'
        '<HASH-AMD64-MSIX>' = 'x64.msix.sha256'
        '<SIG-AMD64-MSIX>' = 'x64.msix.sig.sha256'
        '<HASH-ARM64-MSIX>' = 'arm64.msix.sha256'
        '<SIG-ARM64-MSIX>' = 'arm64.msix.sig.sha256'
    }

    # Collect all installer metadata
    Write-Host "Downloading installer metadata for version $Version..." -ForegroundColor Green
    $metadata = @{ '<VERSION>' = $Version }
    $metadata['<DATE>'] = (Get-Date -Format "yyyy-MM-dd")

    foreach ($placeholder in $metadataFiles.Keys) {
        $cdnFileName = $metadataFiles[$placeholder]
        Write-Verbose "Downloading $placeholder ($cdnFileName)..."
        $metadata[$placeholder] = Get-InstallerMetadata -Architecture $cdnFileName -Version $Version
    }

    # Process all YAML template files
    $yamlFiles = Get-ChildItem -Path '*.yaml' -ErrorAction Stop
    Write-Host "Processing $($yamlFiles.Count) YAML template file(s)..." -ForegroundColor Green

    foreach ($yamlFile in $yamlFiles) {
        Update-ManifestFile -FileName $yamlFile.Name -Metadata $metadata
    }

    Write-Host "Successfully generated manifest files for version $Version" -ForegroundColor Green
}
catch {
    Write-Error "Failed to generate manifest files: $($_.Exception.Message)"
    exit 1
}

#endregion

#region WinGet Submission

# Only proceed with submission if token is provided
if (-not $Token) {
    Write-Host "No GitHub token provided. Skipping WinGet repository submission." -ForegroundColor Yellow
    Write-Host "Manifest files have been generated in the '$Version' directory." -ForegroundColor Green
    exit 0
}

try {
    Write-Host "Preparing to submit to WinGet repository..." -ForegroundColor Green

    # Install the latest wingetcreate exe
    # Need to do things this way, see https://github.com/PowerShell/PowerShell/issues/13138
    Write-Verbose "Importing Appx module using Windows PowerShell compatibility"
    Import-Module Appx -UseWindowsPowerShell -ErrorAction Stop

    # Download and install Winget-Create msixbundle
    $appxBundleFile = Join-Path -Path $env:TEMP -ChildPath "wingetcreate.msixbundle"
    Write-Verbose "Downloading wingetcreate to: $appxBundleFile"

    Invoke-WebRequest -Uri "https://aka.ms/wingetcreate/latest/msixbundle" -OutFile $appxBundleFile -ErrorAction Stop
    Add-AppxPackage -Path $appxBundleFile -ErrorAction Stop

    Write-Verbose "Successfully installed wingetcreate"

    # Submit the PR to WinGet repository
    Write-Host "Submitting pull request to WinGet repository..." -ForegroundColor Green
    wingetcreate submit --token $Token $Version

    if ($LASTEXITCODE -eq 0) {
        Write-Host "Successfully submitted pull request to WinGet repository!" -ForegroundColor Green
    }
    else {
        throw "wingetcreate submit failed with exit code: $LASTEXITCODE"
    }
}
catch {
    Write-Error "Failed to submit to WinGet repository: $($_.Exception.Message)"
    exit 1
}
finally {
    # Clean up temporary files
    if (Test-Path -Path $appxBundleFile) {
        Remove-Item -Path $appxBundleFile -Force -ErrorAction SilentlyContinue
        Write-Verbose "Cleaned up temporary file: $appxBundleFile"
    }
}

#endregion
