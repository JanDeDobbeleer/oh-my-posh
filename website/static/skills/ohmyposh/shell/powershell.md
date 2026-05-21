# Oh My Posh — PowerShell Setup

## Configure

Edit your PowerShell profile. Find its path:

```powershell
$PROFILE
```

Open it (create it if it doesn't exist):

```powershell
notepad $PROFILE
# or: New-Item -Path $PROFILE -Type File -Force
```

Add this as the **last line**:

```powershell
oh-my-posh init pwsh | Invoke-Expression
```

Reload the profile:

```powershell
. $PROFILE
```

## Troubleshooting

**Execution policy blocks scripts:**

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine
```

Or use the `--eval` flag (slightly slower startup):

```powershell
oh-my-posh init pwsh --eval | Invoke-Expression
```

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
