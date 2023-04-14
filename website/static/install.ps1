param(
    [switch]
    $AllUsers
)

$installInstructions = @'
Hey friend

This installer is only available for Windows.
If you're looking for installation instructions for your operating system,
please visit the following link:
'@
if ($IsMacOS) {
    Write-Host @"
$installInstructions

https://ohmyposh.dev/docs/installation/macos
"@
    exit
}
if ($IsLinux) {
    Write-Host @"
$installInstructions

https://ohmyposh.dev/docs/installation/linux
"@
    exit
}

$installer = ''
$arch = (Get-CimInstance -Class Win32_Processor -Property Architecture).Architecture | Select-Object -First 1
switch ($arch) {
    0 { $installer = "install-386.exe" } # x86
    5 { $installer = "install-arm64.exe" } # ARM
    9 {
        if ([Environment]::Is64BitOperatingSystem) {
            $installer = "install-amd64.exe"
        }
        else {
            $installer = "install-386.exe"
        }
    }
    12 { $installer = "install-arm64.exe" } # Surface Pro X
}

if ([string]::IsNullOrEmpty($installer)) {
    Write-Host @"
The installer for system architecture ($arch) is not available.
"@
    exit
}

Write-Host "Downloading $installer..."

# validate the availability of New-TemporaryFile
if (Get-Command -Name New-TemporaryFile -ErrorAction SilentlyContinue) {
    $tmp = New-TemporaryFile | Rename-Item -NewName { $_ -replace 'tmp$', 'exe' } -PassThru
}
else {
    $tmp = New-Item -Path $env:TEMP -Name ([System.IO.Path]::GetRandomFileName() -replace '\.\w+$', '.exe') -Force -ItemType File
}
$url = "https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/$installer"

# check if we can make https requests and download the binary
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $url -Method Head | Where-Object -FilterScript { $_.StatusCode -ne 200 }  # Suppress success output
}
catch {
    Write-Host "Unable to download $installer. Please check your internet connection."
    exit
}

Invoke-WebRequest -OutFile $tmp $url
Write-Host 'Running installer...'
$installMode = "/CURRENTUSER"
if ($AllUsers) {
    $installMode = "/ALLUSERS"
}
& "$tmp" /VERYSILENT $installMode | Out-Null
$tmp | Remove-Item
Write-Host @'
Done!

Restart your terminal and have a look at the
documentation on how to proceed from here.

https://ohmyposh.dev/docs/installation/prompt
'@
