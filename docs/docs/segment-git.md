---
id: git
title: Git
sidebar_label: Git
---

## What

Display `git status` information when in a git repository. Also works for subfolders.
Local changes can also shown by default using the following syntax for both the working and staging area:

- `+` added
- `~` modified
- `-` deleted

## Sample Configuration

```json
{
  "type": "git",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "properties": {
    "branch_icon": "",
    "branch_identical_icon": "≡",
    "branch_ahead_icon": "↑",
    "branch_behind_icon": "↓",
    "local_working_icon": "",
    "local_staged_icon": "",
    "rebase_icon": " ",
    "cherry_pick_icon": " ",
    "commit_icon": " ",
    "tag_icon": "笠 "
  }
}
```

## Properties

- branch_icon: `string` - the icon to use in front of the git branch name
- branch_identical_icon: `string` - the icon to display when remote and local are identical
- branch_ahead_icon: `string` - the icon to display when the local branch is ahead of its remote
- branch_behind_icon: `string` - the icon to display when the local branch is behind its remote
- local_working_icon: `string` - the icon to display in front of the working area changes
- local_staged_icon: `string` - the icon to display in front of the staged area changes
- display_status: `boolean` - display the local changes or not
- rebase_icon: `string` - icon/text to display before the context when in a rebase
- cherry_pick_icon: `string` - icon/text to display before the context when doing a cherry-pick
- commit_icon: `string` - icon/text to display before the commit context
- tag_icon: `string` - icon/text to display before the tag context
