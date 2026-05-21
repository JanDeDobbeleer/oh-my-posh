# Oh My Posh — Windows Installation

## Install

**Recommended (winget):**

```powershell
winget install JanDeDobbeleer.OhMyPosh --source winget
```

**Manual (PowerShell):**

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://ohmyposh.dev/install.ps1'))
```

**Chocolatey:**

```powershell
choco install oh-my-posh
```

## Update

```powershell
winget upgrade JanDeDobbeleer.OhMyPosh --source winget
```

## After installing

Restart the terminal (or open a new window) so `oh-my-posh` is on `$PATH`, then proceed to the
[shell setup guide](/skills/ohmyposh/SKILL.md).

> **Tip:** When using Oh My Posh inside WSL, follow the
> [Linux installation](/skills/ohmyposh/installation/linux.md) guide inside WSL instead.
