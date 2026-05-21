# Oh My Posh — Zsh Setup

## Configure

Add this as the **last line** of `~/.zshrc`:

```bash
eval "$(oh-my-posh init zsh)"
```

Reload:

```bash
exec zsh
```

## macOS Terminal workaround

The standard macOS Terminal has issues with ANSI characters. To skip Oh My Posh there while
keeping it in iTerm2 and other modern terminals:

```bash
if [ "$TERM_PROGRAM" != "Apple_Terminal" ]; then
  eval "$(oh-my-posh init zsh)"
fi
```

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
