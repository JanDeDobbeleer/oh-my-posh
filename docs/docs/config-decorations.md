---
id: config-text-style
title: Text style
sidebar_label: Text style
---

## Text decorations

You can make use of the following syntax to decorate text:

- `<b>bold</b>`: renders `bold` as bold text
- `<u>underline</u>`: renders `underline` as underlined text
- `<i>italic</i>`: renders `italic` as italic text
- `<s>strikethrough</s>`: renders `strikethrough` as strikethrough text

This can be used in templates and icons/text inside your config.

## Hyperlinks

The engine has the ability to render hyperlinks. Your terminal has to support it and the option
has to be enabled at the segment level. Hyperlink generation is disabled by default.

### Supported segments

- [Path][path-segment]
- [Git][git-segment]
- Most languages (version hyperlink)

### Supported terminals

- [Terminal list][terminal-list-hyperlinks]

[terminal-list-hyperlinks]: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
[path-segment]: /docs/path
[git-segment]: /docs/git
