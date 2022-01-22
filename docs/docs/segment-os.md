---
id: os
title: os
sidebar_label: OS
---

## What

Display OS specific info - defaults to Icon.

## Sample Configuration

```json
{
  "type": "os",
  "style": "plain",
  "foreground": "#26C6DA",
  "background": "#546E7A",
  "properties": {
    "postfix": " \uE0B1",
    "macos": "mac"
  }
}
```

## Properties

- macos: `string` - the string to use for macOS - defaults to macOS icon - defaults to `\uF179`
- linux: `string` - the icon to use for Linux - defaults to Linux icon - defaults to `\uF17C`
- windows: `string` - the icon to use for Windows - defaults to Windows icon - defaults to `\uE62A`
- display_distro_name: `boolean` - display the distro name or icon (for WSL and Linux) - defaults to `false`
- alpine: `string` - the icon to use for Alpine - defaults to Alpine icon - defaults to `\uF300`
- aosc: `string` - the icon to use for Aosc - defaults to Aosc icon - defaults to `\uF301`
- arch: `string` - the icon to use for Arch - defaults to Arch icon - defaults to `\uF303`
- centos: `string` - the icon to use for Centos - defaults to Centos icon - defaults to `\uF304`
- coreos: `string` - the icon to use for Coreos - defaults to Coreos icon - defaults to `\uF305`
- debian: `string` - the icon to use for Debian - defaults to Debian icon - defaults to `\uF306`
- devuan: `string` - the icon to use for Devuan - defaults to Devuan icon - defaults to `\uF307`
- raspbian: `string` - the icon to use for Raspbian - defaults to Raspbian icon - defaults to `\uF315`
- elementary: `string` - the icon to use for Elementary - defaults to Elementary icon - defaults to `\uF309`
- fedora: `string` - the icon to use for Fedora - defaults to Fedora icon - defaults to `\uF30a`
- gentoo: `string` - the icon to use for Gentoo - defaults to Gentoo icon - defaults to `\uF30d`
- mageia: `string` - the icon to use for Mageia - defaults to Mageia icon - defaults to `\uF310`
- manjaro: `string` - the icon to use for Manjaro - defaults to Manjaro icon - defaults to `\uF312`
- mint: `string` - the icon to use for Mint - defaults to Mint icon - defaults to `\uF30e`
- nixos: `string` - the icon to use for Nixos - defaults to Nixos icon - defaults to `\uF313`
- opensuse: `string` - the icon to use for Opensuse - defaults to Opensuse icon - defaults to `\uF314`
- sabayon: `string` - the icon to use for Sabayon - defaults to Sabayon icon - defaults to `\uF317`
- slackware: `string` - the icon to use for Slackware - defaults to Slackware icon - defaults to `\uF319`
- ubuntu: `string` - the icon to use for Ubuntu - defaults to Ubuntu icon - defaults to `\uF31b`

## [Template][templates] Properties

- `.OS`: `string` - the OS platform

[templates]: /docs/config-text#templates
