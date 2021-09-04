---
id: memory
title: Memory
sidebar_label: Memory
---

## Memory

Display physical memory or swap usage percentage.

## Sample Configuration

```json
{
  "type": "memory",
  "style": "powerline",
  "foreground": "#26C6DA",
  "background": "#2f2f2f",
  "properties": {
    "precision": 1,
    "prefix": " \uf85a ",
    "postfix": "% "
  }
}
```

## Properties

- precision: `int` - the number of decimal places to show - defaults to `0`
- use_available: `boolean` - whether to use available or free memory on Linux - defaults to `true`
- memory_type: `enum` - whether to show `physical` memory or `swap` memory - defaults to `physical`
