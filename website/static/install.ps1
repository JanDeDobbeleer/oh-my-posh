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
$arch = (Get-CimInstance -Class Win32_Processor -Property Architecture).Architecture
switch ($arch) {
    0 { $installer = "install-386.exe" } # x86
    5 { $installer = "install-arm64.exe" } # ARM
    9 {
        if ([Environment]::Is64BitOperatingSystem) {
            $installer = "install-amd64.exe"
        } else {
            $installer = "install-386.exe"
        }
    }
    12 { $installer = "install-arm64.exe" } # Surface Pro X
}

if ($installer -eq '') {
    Write-Host @"
The installer for system architecture ($arch) is not available.
"@
    exit
}

Write-Host "Downloading $installer..."
$tmp = New-TemporaryFile | Rename-Item -NewName { $_ -replace 'tmp$', 'exe' } -PassThru
$url = "https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/$installer"
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
