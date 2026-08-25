# MSIX Package

## Prerequisites

- [Windows SDK] (provides `makeappx.exe` and `signtool.exe`)

## Build the package

This guide assumes and advises the use of PowerShell as your shell environment for this purpose.

Place the executable to package at `./dist/oh-my-posh.exe`, or use `-Copy` to take it from
the repository's `dist` folder.

```powershell
./build.ps1 -Architecture x64 -Version "1.3.37"
```

The package is created at `out/install-x64.msix`. The architecture-specific App Installer feed is
created at `out/install-x64.appinstaller` and embedded in the MSIX as `Update.appinstaller`.

## Install the package

```powershell
Add-AppxPackage -Path ./out/install-x64.msix
```

On Windows 11 build 22000 and later, direct MSIX installations register the embedded App Installer
feed and check for updates on launch when the previous check was at least 24 hours ago. Earlier
supported Windows versions ignore the update metadata and install the package normally.

[Windows SDK]: https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
