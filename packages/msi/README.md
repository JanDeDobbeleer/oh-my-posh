# MSI Package

## Prerequisites

- [dotnet]
- [wix]: `dotnet tool install --global wix`

## Build the package

This guide assumes and advices the use of PowerShell as your shell environment for this purpose.

### Set the environment variables

```powershell
$env:VERSION = "1.3.37"
```

### Build the installer

```powershell
wix build -arch arm64 -out install-arm64.msi
```

## Install the package

### For the current user

```powershell
install-arm64.msi
```

### For all users

```powershell
install-arm64.msi ALLUSERS=1
```

[dotnet]: https://dotnet.microsoft.com/en-us/download/dotnet?cid=getdotnetcorecli
[wix]: https://wixtoolset.org/docs/intro/
