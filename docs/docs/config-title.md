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

### Console Title Template

The following examples illustrate possible contents for `console_title_template`, provided
the current working directory is `/usr/home/omp` and the shell is `zsh`.

To learn more about templates and their possibilities, have a look at the [template][templates] section.

```json
{
    "console_title_template": "{{.Folder}}{{if .Root}} :: root{{end}} :: {{.Shell}}",
    // outputs:
    // when root == false: omp :: zsh
    // when root == true: omp :: root :: zsh
    "console_title_template": "{{.Folder}}", // outputs: omp
    "console_title_template": "{{.Shell}} in {{.PWD}}", // outputs: zsh in /usr/home/omp
    "console_title_template": "{{.UserName}}@{{.HostName}} {{.Shell}} in {{.PWD}}", // outputs: MyUser@MyMachine zsh in /usr/home/omp
    "console_title_template": "{{.Env.USERDOMAIN}} {{.Shell}} in {{.PWD}}", // outputs: MyCompany zsh in /usr/home/omp
}
```

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[templates]: /docs/config-templates
