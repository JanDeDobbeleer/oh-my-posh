---
name: oh-my-posh
description: "Install, configure, or troubleshoot Oh My Posh/ohmyposh: shell init, themes, segments, Nerd Font icons, and prompt setup on PowerShell, zsh, bash, or fish."
---

# Oh My Posh

> Oh My Posh is a cross-shell, cross-platform prompt theme engine. It displays Git status,
> language versions, cloud context, system metrics, and 180+ other segments in a beautifully
> themed terminal prompt.

## How to use this skill

1. **Identify the user's OS** → follow the matching installation guide:
   - [Windows installation](/skills/ohmyposh/installation/windows.md)
   - [macOS installation](/skills/ohmyposh/installation/macos.md)
   - [Linux installation](/skills/ohmyposh/installation/linux.md)

2. **Identify the user's shell** → follow the matching shell setup guide:
   - [PowerShell](/skills/ohmyposh/shell/powershell.md)
   - [Bash](/skills/ohmyposh/shell/bash.md)
   - [Zsh](/skills/ohmyposh/shell/zsh.md)
   - [Fish](/skills/ohmyposh/shell/fish.md)
   - [Nu / Nushell](/skills/ohmyposh/shell/nu.md)
   - [Cmd (via Clink)](/skills/ohmyposh/shell/cmd.md)
   - [Elvish](/skills/ohmyposh/shell/elvish.md)
   - [Xonsh](/skills/ohmyposh/shell/xonsh.md)

3. **Install a Nerd Font** — required for icons and glyphs:

   ```bash
   oh-my-posh font install
   ```

   Recommended: **Meslo LGM NF**. Set it in the terminal emulator's font settings after installing.

4. **Customize the prompt** → see [configuration](/skills/ohmyposh/configuration.md)

## Detect the current shell

If the user is unsure which shell they are using (requires oh-my-posh to be installed first):

```bash
oh-my-posh get shell
```

## Troubleshooting

- **Icons show as boxes** — install a Nerd Font and set it in the terminal emulator.
- **`oh-my-posh` not found after install** — restart the terminal or add the install path to `$PATH`.
- **Slow prompt** -- enable async rendering by setting `"async": true` at the top level of your config.

Full documentation: <https://ohmyposh.dev/docs>
FAQ: <https://ohmyposh.dev/docs/faq>
