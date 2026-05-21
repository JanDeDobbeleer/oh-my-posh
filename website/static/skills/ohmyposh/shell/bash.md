# Oh My Posh — Bash Setup

## Configure

Add this as the **last line** of `~/.bashrc` (or `~/.profile` / `~/.bash_profile` depending on
your environment):

```bash
eval "$(oh-my-posh init bash)"
```

Reload:

```bash
exec bash
```

Or, if using `~/.profile`:

```bash
. ~/.profile
```

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
