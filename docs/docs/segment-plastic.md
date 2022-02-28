---
id: plastic
title: Plastic SCM
sidebar_label: Plastic SCM
---

## What

Display Plastic SCM information when in a plastic repository. Also works for subfolders.
For maximum compatibility, make sure your `cm` executable is up-to-date
(when branch or status information is incorrect for example).

Local changes can also be displayed which uses the following syntax (see `.Status` property below):

- `+` added
- `~` modified
- `-` deleted
- `>` moved
- `x` unmerged

## Sample Configuration

```json
{
  "type": "plastic",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "background_templates": [
    "{{ if .MergePending }}#006060{{ end }}",
    "{{ if .Changed }}#FF9248{{ end }}",
    "{{ if and .Changed .Behind }}#ff4500{{ end }}",
    "{{ if .Behind }}#B388FF{{ end }}"
  ],
  "properties": {
    "fetch_status": true,
    "branch_max_length": 25,
    "truncate_symbol": "\u2026",
    "template": "{{ .Selector }}{{ if .Status.Changed }} \uF044 {{ end }}{{ .Status.String }}"
  }
}
```

## Plastic SCM Icon

If you want to use the icon of Plastic SCM in the segment, then please help me push the icon in this [issue][fa-issue]
by leaving a like!
![icon](https://www.plasticscm.com/images/icon-logo-plasticscm.svg)

## Properties

### Fetching information

As doing multiple `cm` calls can slow down the prompt experience, we do not fetch information by default.
You can set the following property to `true` to enable fetching additional information (and populate the template).

- fetch_status: `boolean` - fetch the local changes - defaults to `false`

### Icons

#### Branch

- branch_icon: `string` - the icon to use in front of the git branch name - defaults to `\uE0A0 `
- full_branch_path: `bool` - display the full branch path: */main/fix-001* instead of *fix-001* - defaults to `true`
- branch_max_length: `int` - the max length for the displayed branch name where `0` implies full length - defaults to `0`
- truncate_symbol: `string` - the icon to display when a branch name is truncated - defaults to empty

#### Selector

- commit_icon: `string` - icon/text to display before the commit context (detached HEAD) - defaults to `\uF417`
- tag_icon: `string` - icon/text to display before the tag context - defaults to `\uF412`

## Template ([info][templates])

:::note default template

``` template
{{ .Selector }}
```

:::

### Properties

- `.Selector`: `string` - the current selector context (branch/changeset/label)
- `.Behind`: `bool` - the current workspace is behind and changes are incoming
- `.Status`: `PlasticStatus` - changes in the workspace (see below)
- `.MergePending`: `bool` - if a merge is pending and needs to be commited
(kown issue: when no file is left after a *Change/Delete conflict* merge, the `MergePending` property is not set)

### PlasticStatus

- `.Unmerged`: `int` - number of unmerged changes
- `.Deleted`: `int` - number of deleted changes
- `.Added`: `int` - number of added changes
- `.Modified`: `int` - number of modified changes
- `.Moved`: `int` - number of moved changes
- `.Changed`: `boolean` - if the status contains changes or not
- `.String`: `string` - a string representation of the changes above

[templates]: /docs/config-templates
[fa-issue]: https://github.com/FortAwesome/Font-Awesome/issues/18504
