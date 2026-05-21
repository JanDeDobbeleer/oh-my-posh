---
name: oh-my-posh
description: Install and configure Oh My Posh — the most customizable prompt theme engine for any shell — on Windows, macOS, or Linux. Guides users step-by-step through installation and shell integration for PowerShell, Bash, Zsh, Fish, Nu, Cmd, Elvish, and Xonsh.
---

# Oh My Posh Setup Skill

Use this skill to help a user install and configure Oh My Posh in their terminal.

## Step 1 — Identify OS and shell

Ask the user:
1. Which **operating system** they are on: Windows, macOS, or Linux.
2. Which **shell** they want to configure: PowerShell (pwsh), Bash, Zsh, Fish, Nu (Nushell), Cmd, Elvish, or Xonsh.

If they are unsure of their shell, tell them to run:

```bash
oh-my-posh get shell
```

(They must have Oh My Posh installed first — see Step 2.)

---

## Step 2 — Install Oh My Posh

### Windows

**Recommended (winget):**
```powershell
winget install JanDeDobbeleer.OhMyPosh --source winget
```

**Alternative (manual, from PowerShell):**
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://ohmyposh.dev/install.ps1'))
```

**Alternative (Chocolatey):**
```powershell
choco install oh-my-posh
```

### macOS

**Recommended (Homebrew):**
```bash
brew install jandedobbeleer/oh-my-posh/oh-my-posh
```

**Manual:**
```bash
curl -s https://ohmyposh.dev/install.sh | bash -s
```

### Linux

```bash
curl -s https://ohmyposh.dev/install.sh | bash -s
```

---

## Step 3 — Install a Nerd Font (recommended)

Oh My Posh themes use icon glyphs from Nerd Fonts. Without a Nerd Font, prompts will show
placeholder boxes instead of icons.

Install a Nerd Font with:
```bash
oh-my-posh font install
```

Then set the installed font in your terminal emulator's settings (e.g., Windows Terminal, iTerm2,
GNOME Terminal). A recommended font is **Meslo LGM NF**.

---

## Step 4 — Configure your shell

Add the appropriate init snippet to the end of your shell's profile or rc file, then reload.

### PowerShell (`$PROFILE`)
```powershell
oh-my-posh init pwsh | Invoke-Expression
```
Reload with: `. $PROFILE`

> If the profile file doesn't exist yet: `New-Item -Path $PROFILE -Type File -Force`
> If scripts are blocked: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine`

### Bash (`~/.bashrc`)
```bash
eval "$(oh-my-posh init bash)"
```
Reload with: `exec bash`

### Zsh (`~/.zshrc`)
```bash
eval "$(oh-my-posh init zsh)"
```
Reload with: `exec zsh`

### Fish (`~/.config/fish/config.fish`)
```bash
oh-my-posh init fish | source
```
Reload with: `exec fish`

### Nu / Nushell (`$nu.config-path`)
```bash
oh-my-posh init nu
```
Restart Nushell to apply.

### Cmd (via Clink — `oh-my-posh.lua` in Clink scripts directory)
```lua
load(io.popen('oh-my-posh init cmd'):read("*a"))()
```
Install [Clink](https://chrisant996.github.io/clink/) first if needed, then restart Cmd.

### Elvish (`~/.elvish/rc.elv`)
```bash
eval (oh-my-posh init elvish)
```
Reload with: `exec elvish`

### Xonsh (`~/.xonshrc`)
```bash
execx($(oh-my-posh init xonsh))
```
Reload with: `exec xonsh`

---

## Step 5 — Choose a theme (optional)

List and preview bundled themes:
```bash
oh-my-posh debug
```

Browse themes at: https://ohmyposh.dev/docs/themes

Apply a theme by passing `--config` to the init command, for example:
```bash
# PowerShell
oh-my-posh init pwsh --config "$env:POSH_THEMES_PATH\jandedobbeleer.omp.json" | Invoke-Expression

# Bash/Zsh
eval "$(oh-my-posh init bash --config $(brew --prefix oh-my-posh)/themes/jandedobbeleer.omp.json)"
```

The themes directory location is stored in `$POSH_THEMES_PATH` (Windows) or can be found with
`oh-my-posh env` after installation.

---

## Step 6 — Validate a custom config (optional)

If the user has a custom config file, validate it using the MCP endpoint:

```bash
curl -X POST https://ohmyposh.dev/api/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"validate_config","arguments":{"content":"<CONFIG_CONTENT>","format":"auto"}},"id":1}'
```

---

## Troubleshooting

- **Icons show as boxes**: Install a Nerd Font (Step 3) and set it in your terminal emulator.
- **`oh-my-posh` not found after install**: Restart your terminal or add the install location to `$PATH`.
- **PowerShell execution policy error**: Run `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope LocalMachine`.
- **Fish version too old**: Upgrade to Fish 4.1.0+ for full support including transient prompt.
- **Nu version too old**: Oh My Posh requires Nushell v0.104.0 or higher.

Full documentation: https://ohmyposh.dev/docs
