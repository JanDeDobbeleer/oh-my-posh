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
    "template": "{{ .UserName }}"
  }
}
```

## Properties

- ssh_icon: `string` - text/icon to display first when in an active SSH session - defaults
to `\uF817 `
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below.

## Template Properties

- `.UserName`: `string` - the current user's name
- `.ComputerName`: `string` - the current computer's name
- `.SSHSession`: `boolean` - active SSH session or not
- `.Root`: `boolean` - are you a root/admin user or not

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
