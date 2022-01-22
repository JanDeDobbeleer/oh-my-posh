---
id: session
title: Session
sidebar_label: Session
---

## What

Show the current user and host name.

## Sample Configuration

```json
{
  "type": "session",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#c386f1",
  "leading_diamond": "\uE0B6",
  "trailing_diamond": "\uE0B0",
  "properties": {
    "template": "{{ if .SSHSession }}\uF817 {{ end }}{{ .UserName }}"
  }
}
```

## [Template][templates] Properties

- `.UserName`: `string` - the current user's name
- `.HostName`: `string` - the current computer's name
- `.SSHSession`: `boolean` - active SSH session or not
- `.Root`: `boolean` - are you a root/admin user or not

[templates]: /docs/config-text#templates
