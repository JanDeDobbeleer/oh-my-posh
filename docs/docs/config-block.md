---
id: config-block
title: Block
sidebar_label: Block
---

Let's take a closer look at what defines a block.

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  ...
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        ...
      ]
    }
  ]
}
```

- type: `prompt` | `rprompt`
- newline: `boolean`
- alignment: `left` | `right`
- filler: `string`
- segments: `array` of one or more `segments`

### Type

Tells the engine what to do with the block. There are two options:

- `prompt` renders one or more segments
- `rprompt` renders one or more segments aligned to the right of the cursor. Only one `rprompt` block is permitted.
Supported on zsh, bash, PowerShell, cmd and fish.

### Newline

Start the block on a new line - defaults to `false`.

### Alignment

Tell the engine if the block should be left or right-aligned.

### Filler

When you want to join a right and left aligned block with a repeated set of characters, add the character
to be repeated to this property. Add this property to the _right_ aligned block.

```json
"alignment": "right",
"filler": "."
```

### Segments

Array of one or more segments.
