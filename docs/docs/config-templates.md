---
id: config-templates
title: Templates
sidebar_label: Templates
---

Every segment has a `template` property to tweak the text that is displayed.
Under the hood, this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with.

## Global properties

- `.Root`: `boolean` - is the current user root/admin or not
- `.PWD`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

## Helper functions

- url: create a hyperlink to a website to open your default browser `{{ url .UpstreamIcon .UpstreamURL }}`
(needs terminal [support][terminal-list-hyperlinks])
- path: create a link to a folder to open your file explorer `{{ path .Path .Location }}`
(needs terminal [support][terminal-list-hyperlinks])
- secondsRound: round seconds to a time indication `{{ secondsRound 3600 }}` -> 1h

## Text decoration

You can make use of the following syntax to decorate text:

- `<b>bold</b>`: renders `bold` as bold text
- `<u>underline</u>`: renders `underline` as underlined text
- `<i>italic</i>`: renders `italic` as italic text
- `<s>strikethrough</s>`: renders `strikethrough` as strikethrough text

This can be used in templates and icons/text inside your config.

[terminal-list-hyperlinks]: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
[path-segment]: /docs/path
[git-segment]: /docs/git
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
