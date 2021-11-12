---
id: config-title
title: Console title
sidebar_label: Console title
---

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "console_title_template": "{{.Folder}}{{if .Root}} :: root{{end}} :: {{.Shell}}",
  "blocks": [
    ...
  ]
}
```

To manipulate the console title, you can make use of the following properties:

- console_title: `boolean` - when true sets the current location as the console title
- console_title_style: `string` - the title to set in the console - defaults to `folder`
- console_title_template: `string` - the template to use when `"console_title_style" = "template"`

### Console Title Style

- `folder`: show the current folder name
- `path`: show the current path
- `template`: show a custom template

### Console Title Template

You can create a more custom console title with the use of `"console_title_style" = "template"`.
When this is set, a `console_title_template` is also expected, otherwise, the title will remain empty.
Under the hood, this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with.

- `.Root`: `boolean` - is the current user root/admin or not
- `.Path`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.User`: `string` - the current user name
- `.Host`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

A `boolean` can be used for conditional display purposes, a `string` can be displayed.

The following examples illustrate possible contents for `console_title_template`, provided
the current working directory is `/usr/home/omp` and the shell is `zsh`.

```json
{
    "console_title_template": "{{.Folder}}{{if .Root}} :: root{{end}} :: {{.Shell}}",
    // outputs:
    // when root == false: omp :: zsh
    // when root == true: omp :: root :: zsh
    "console_title_template": "{{.Folder}}", // outputs: omp
    "console_title_template": "{{.Shell}} in {{.Path}}", // outputs: zsh in /usr/home/omp
    "console_title_template": "{{.User}}@{{.Host}} {{.Shell}} in {{.Path}}", // outputs: MyUser@MyMachine zsh in /usr/home/omp
    "console_title_template": "{{.Env.USERDOMAIN}} {{.Shell}} in {{.Path}}", // outputs: MyCompany zsh in /usr/home/omp
}
```

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
