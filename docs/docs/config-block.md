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
- vertical_offset: `int`
- horizontal_offset: `int`
- segments: `array` of one or more `segments`

### Type

Tells the engine what to do with the block. There are three options:

- `prompt` renders one or more segments
- `rprompt` renders one or more segments aligned to the right of the cursor. Only one `rprompt` block is permitted.
Supported on [ZSH][rprompt], Bash and Powershell.

### Newline

Start the block on a new line. Defaults to `false`.

### Alignment

Tell the engine if the block should be left or right-aligned.

### Vertical offset

Move the block up or down x lines. For example, `vertical_offset: 1` moves the prompt down one line, `vertical_offset: -1`
moves it up one line.

### Horizontal offset

Moves the segment to the left or the right to have it exactly where you want it to be. Works like `vertical_offset`
but on a horizontal level where a negative number moves the block left and a positive number right.

### Segments

Array of one or more segments.
