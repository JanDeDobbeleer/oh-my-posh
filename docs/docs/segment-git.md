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
  "powerline_symbol": "\uE0B0",
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
    "tag_icon": "笠 ",
    "display_stash_count": true,
    "stash_count_icon": "\uF692 ",
    "merge_icon": "\uE726 ",
    "display_upstream_icon": true,
    "github_icon": "\uE709",
    "bitbucket_icon": "\uE703",
    "gitlab_icon": "\uE296",
    "git_icon": "\uE702"
  }
}
```

## Properties

### Standard

- branch_icon: `string` - the icon to use in front of the git branch name
- branch_identical_icon: `string` - the icon to display when remote and local are identical
- branch_ahead_icon: `string` - the icon to display when the local branch is ahead of its remote
- branch_behind_icon: `string` - the icon to display when the local branch is behind its remote

### Status

- display_status: `boolean` - display the local changes or not
- display_stash_count: `boolean` show stash count or not
- status_separator_icon: `string` icon/text to display between staging and working area changes
- local_working_icon: `string` - the icon to display in front of the working area changes
- local_staged_icon: `string` - the icon to display in front of the staged area changes
- stash_count_icon: `string` icon/text to display before the stash context

### HEAD context

- commit_icon: `string` - icon/text to display before the commit context (detached HEAD)
- tag_icon: `string` - icon/text to display before the tag context
- rebase_icon: `string` - icon/text to display before the context when in a rebase
- cherry_pick_icon: `string` - icon/text to display before the context when doing a cherry-pick
- merge_icon: `string` icon/text to display before the merge context

### Upstream context

- display_upstream_icon: `boolean` - display upstrean icon or not
- github_icon: `string` - icon/text to display when the upstream is Github
- gitlab_icon: `string` - icon/text to display when the upstream is Gitlab
- bitbucket_icon: `string` - icon/text to display when the upstream is Bitbucket
- git_icon: `string` - icon/text to display when the upstream is not known/mapped
