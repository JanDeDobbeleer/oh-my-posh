---
id: kubectl
title: Kubectl Context
sidebar_label: Kubectl
---

## What

Display the currently active Kubernetes context name and namespace name.

## Sample Configuration

```json
{
  "type": "kubectl",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#000000",
  "background": "#ebcc34",
  "properties": {
    "prefix": " \uFD31 ",
    "template": "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}"
  }
}
```

## Properties

- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}`
- display_error: `boolean` - show the error context when failing to retrieve the kubectl information - defaults to `false`

## Template Properties

- `.Context`: `string` - the current kubectl context
- `.Namespace`: `string` - the current kubectl namespace

## Tips

It is common for the Kubernetes "default" namespace to be used when no namespace is provided. If you want your prompt to
 render an empty current namespace using the word "default", you can use something like this for the template:

`{{.Context}} :: {{if .Namespace}}{{.Namespace}}{{else}}default{{end}}`

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
