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
file at `out/install-x64.appinstaller`.

The `.appinstaller` file references the published package on the CDN and enables
automatic updates: Windows re-checks the stable URL it was installed from
(daily on launch and via a background task) and updates the package silently
when a new version is available.

## Install the package

```powershell
Add-AppxPackage -Path ./out/install-x64.msix
```

## Install via App Installer (with auto-update)

```powershell
Add-AppxPackage -AppInstallerFile https://cdn.ohmyposh.dev/releases/latest/install-x64.appinstaller
```

[Windows SDK]: https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
