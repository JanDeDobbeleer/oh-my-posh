# Oh My Posh — Cmd Setup (via Clink)

Cmd does not support custom prompts natively. [Clink](https://chrisant996.github.io/clink/)
adds that capability and also enhances Cmd with readline-style editing.

## Prerequisites

Install Clink and enable autostart. Find the Clink scripts directory:

```bash
clink info
```

## Configure

Create a file called `oh-my-posh.lua` in the Clink scripts directory:

```lua
load(io.popen('oh-my-posh init cmd'):read("*a"))()
```

Restart Cmd to apply.

> **Tip:** Clink has built-in support for Oh My Posh. You can also set the prompt with:
>
> ```bash
> clink config prompt use oh-my-posh
> ```

## Next step

→ [Customize your prompt](/skills/ohmyposh/configuration.md)
