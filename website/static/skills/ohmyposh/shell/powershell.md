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
$(if (Get-Command 'oh-my-posh' -ErrorAction Ignore) {
    oh-my-posh init pwsh
    # Set output encoding to UTF-8
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    # Set input encoding to UTF-8 (for reading user input with non-ASCII chars)
    [Console]::InputEncoding = [System.Text.Encoding]::UTF8
}) | Invoke-Expression
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

Or add the `--eval` flag to the `oh-my-posh init pwsh` command above (slightly slower startup).

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
