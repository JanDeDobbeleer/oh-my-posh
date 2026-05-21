# Oh My Posh — macOS Installation

> **Note:** The standard macOS Terminal only supports 256 colors. Use [iTerm2](https://iterm2.com)
> or another modern terminal for full ANSI color and icon support.

## Install

**Recommended (Homebrew):**

```bash
brew install jandedobbeleer/oh-my-posh/oh-my-posh
```

**MacPorts** (community-maintained):

```bash
sudo port selfupdate
sudo port install oh-my-posh
```

## Update

```bash
brew upgrade oh-my-posh
```

## After installing

Restart the terminal (or open a new window) so `oh-my-posh` is on `$PATH`, then proceed to the
[shell setup guide](/skills/ohmyposh/ohmyposh.md).

## Troubleshooting

If you see "conditional binary operator expected", your system bash is outdated. Update it:

```bash
brew install bash
grep -qxF "$(brew --prefix)/bin/bash" /etc/shells || sudo bash -c 'echo "$(brew --prefix)/bin/bash" >> /etc/shells'
chsh -s "$(brew --prefix)/bin/bash" $USER
```
