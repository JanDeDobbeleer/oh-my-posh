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

## Properties

- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below.

## Template Properties

- `.UserName`: `string` - the current user's name
- `.ComputerName`: `string` - the current computer's name
- `.SSHSession`: `boolean` - active SSH session or not
- `.Root`: `boolean` - are you a root/admin user or not

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
