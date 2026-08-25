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

The package is created at `out/install-x64.msix`, along with an App Installer
file at `out/install-x64.appinstaller`. The App Installer configuration is also
embedded in the MSIX, so installing the MSIX enables automatic updates on
supported Windows versions.

## Install the package

```powershell
Add-AppxPackage -Path ./out/install-x64.msix
```

[Windows SDK]: https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
