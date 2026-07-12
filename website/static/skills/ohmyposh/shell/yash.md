# Oh My Posh — Yash Setup

## Configure

Add this as the **last line** of `~/.yashrc`:

```bash
eval "$(oh-my-posh init yash)"
```

Reload:

```bash
exec yash
```

## Notes

- Requires a yash version that supports `$PROMPT_COMMAND` (>= 2.39) and `$POST_PROMPT_COMMAND`
  (>= 2.52) for full functionality; oh-my-posh hooks into both to time commands and draw the prompt.
- The prompt is rendered through `YASH_PS1` (and `YASH_PS1R` for the right prompt), which take
  precedence over `PS1`/`PS2` in yash.
- Right prompt support is native via `YASH_PS1R`.
- Transient prompt and tooltips are not supported in yash.

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
