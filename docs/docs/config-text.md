---
id: config-text
title: Text
sidebar_label: Text
---

## Templates

Every segment has a `template` property to tweak the text that is displayed.
Under the hood, this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with.

- `.Root`: `boolean` - is the current user root/admin or not
- `.PWD`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

## Text decoration

You can make use of the following syntax to decorate text:

- `<b>bold</b>`: renders `bold` as bold text
- `<u>underline</u>`: renders `underline` as underlined text
- `<i>italic</i>`: renders `italic` as italic text
- `<s>strikethrough</s>`: renders `strikethrough` as strikethrough text

This can be used in templates and icons/text inside your config.

## Hyperlinks

The engine has the ability to render hyperlinks. Your terminal has to support it and the option
has to be enabled at the segment level. Hyperlink generation is disabled by default.

When using in a template, the syntax is like markdown:

- url: `[text](https://link)`
- file: `[file](file://path)`

### Supported segments

- [Path][path-segment]
- [Git][git-segment]
- Most languages (version hyperlink)

### Supported terminals

- [Terminal list][terminal-list-hyperlinks]

[terminal-list-hyperlinks]: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
[path-segment]: /docs/path
[git-segment]: /docs/git
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
