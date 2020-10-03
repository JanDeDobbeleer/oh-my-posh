# PowerShell package

The goal of this module is to wrap the `oh-my-posh` binaries into a PowerShell module and allow easy installation
and ease of use when setting the prompt.

## Testing

## Create local package repository to validate changes locally

### Create the repository

```powershell
Register-PSRepository -Name 'LocalRepo' -SourceLocation 'C:\Repo' -PublishLocation 'C:\Repo' -InstallationPolicy Trusted
```

## Make changes and publish to your local repository

For ease testing, up the version number when using the build script, that way you are always able to update the module.

```powershell
deploy.ps1 -BinVersion 0.1.0 -ModuleVersion 0.0.2 -Repository LocalRepo
```

## Validate changes

Install/Update the module from your local repository and validate the changes.

```powershell
Install-Module oh-my-posh -Repository LocalRepo -Force
```
