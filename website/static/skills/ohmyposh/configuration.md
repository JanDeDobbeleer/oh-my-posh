# Oh My Posh — Configuration & Customization

## Apply a theme

Pass `--config` to the init command with a theme name, local path, or remote URL.

**By theme name** (no extension needed):

```powershell
# PowerShell
oh-my-posh init pwsh --config 'jandedobbeleer' | Invoke-Expression
```

```bash
# Bash / Zsh
eval "$(oh-my-posh init bash --config jandedobbeleer)"
```

**By local file path:**

```powershell
# PowerShell
oh-my-posh init pwsh --config 'C:\Users\<YourUsername>\.mytheme.omp.json' | Invoke-Expression
```

```bash
# Bash / Zsh
eval "$(oh-my-posh init bash --config ~/.mytheme.omp.json)"
```

**By remote URL:**

```powershell
oh-my-posh init pwsh --config 'https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/jandedobbeleer.omp.json' | Invoke-Expression
```

> Using a remote URL adds a network dependency. Oh My Posh caches remote configs using ETags,
> but latency still applies on cache misses. For reliable offline use, copy the theme to a
> local file and use a local path instead.

## Browse available themes

View all bundled themes online: <https://ohmyposh.dev/docs/themes>

> There is no terminal command to preview all available themes. Direct users to the website above.

## Debug the current theme

To render your active prompt in debug mode (shows segment timing and values):

```bash
oh-my-posh debug
```

## Export a theme for editing

```bash
oh-my-posh config export --config jandedobbeleer --output ~/.mytheme.omp.json
```

Then set `--config ~/.mytheme.omp.json` in your init line and edit the file to taste.

## WSL tip

Inside WSL you can share a theme stored in your Windows home folder:

```bash
eval "$(oh-my-posh init bash --config /mnt/c/Users/<WINDOWSUSERNAME>/mytheme.omp.json)"
```

## Live reload during editing

Enable live reload so prompt changes appear without restarting the shell:

```bash
oh-my-posh enable reload
```

Disable:

```bash
oh-my-posh disable reload
```

Preview every configured prompt without restarting:

```bash
oh-my-posh print preview
```

Force-render all segments regardless of context:

```bash
oh-my-posh print preview --force
```

## Validate a config or segment via MCP

When creating or editing a segment manually, use the oh-my-posh MCP server to validate the
config or a segment snippet for schema errors. Agents can call:

```bash
curl -X POST https://ohmyposh.dev/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "validate_config",
      "arguments": {"content": "<CONFIG_CONTENT>", "format": "auto"}
    },
    "id": 1
  }'
```

Or use `validate_segment` to check a single segment snippet instead of an entire config.

## Configuration reference

- Full schema: <https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json>
- Docs: <https://ohmyposh.dev/docs/configuration/overview>
- Segments: <https://ohmyposh.dev/docs/segments/overview>
