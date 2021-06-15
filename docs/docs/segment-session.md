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
  "trailing_diamond": "\uE0B0"
}
```

## Properties

- user_info_separator: `string` - text/icon to put in between the user and host name - defaults to `@`
- ssh_icon: `string` - text/icon to display first when in an active SSH session - defaults
to `\uF817 `
- user_color: `string` [color][colors] - override the foreground color of the user name
- host_color: `string` [color][colors] - override the foreground color of the host name
- display_user: `boolean` - display the user name or not - defaults to `true`
- display_host: `boolean` - display the host name or not - defaults to `true`
- default_user_name: `string` - name of the default user - defaults to empty
- display_default: `boolean` - display the segment or not when the user matches `default_user_name` - defaults
to `true`
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Only used when a value is set, making the above properties obsolete.

## Template Properties

- `.UserName`: `string` - the current user's name
- `.DefaultUserName`: - the default user name (set with the `POSH_SESSION_DEFAULT_USER` env var or `default_user_name` property)
- `.ComputerName`: `string` - the current computer's name
- `.SSHSession`: `boolean` - active SSH session or not
- `.Root`: `boolean` - are you a root/admin user or not

## Environmnent Variables

- `POSH_SESSION_DEFAULT_USER` - used to override the hardcoded `default_user_name` property

[colors]: /docs/configure#colors
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
