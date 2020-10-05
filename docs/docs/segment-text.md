---
id: text
title: Text
sidebar_label: Text
---

## What

Display text.

## Sample Configuration

```json
{
  "type": "text",
  "style": "plain",
  "foreground": "#E06C75",
  "properties": {
    "prefix": "",
    "text": " ❯"
  }
}
```

## Properties

- text: `string` - text/icon to display. Accepts [coloring foreground][coloring] just like `prefix` and `postfix`.

[coloring]: /docs/configure#colors
