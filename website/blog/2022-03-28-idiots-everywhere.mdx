---
title: "What idiot wrote this code?"
description: "It was me..."
slug: idiots-everywhere
authors:
- name: Jan De Dobbeleer
  title: Maintainer
  url: https://github.com/jandedobbeleer
  image_url: https://avatars.githubusercontent.com/u/2492783?v=4
tags: [weekly, ohmyposh]
hide_table_of_contents: false
---

It finally happened, **I introduced a bug** that only appears in 1 use case and it can’t be resolved as it’s the result of
using an old and new version of Oh My Posh at the same time. Who does that you might ask? Well, it can be because you’re
using WSL on Windows, or VM’s on any system while sharing the same configuration cross installation.
Not the main use-case, but also not exotic either.

## What happened?

In the beginning of the year, after what has been quite the struggle, I was finally able to deliver **the most important
architectural update** since the introduction of the rewrite in Go. Because of this, a migration of old configs was needed
so that no end user would have to go through the ordeal of figuring out what to map where, it's a very logical process
and obviously computers are better at that.

To facilitate this, I added a config property called `version` which can be incremented when the configuration evolves
and triggers an automatic migration when necessary. The logic to identify the need for migration is as follows:

```go
if !env.Flags().Migrate && cfg.Version != configVersion {
    cfg.BackupAndMigrate(env)
}
```

If you're observant, you might already spot the issue now that you know the use-case. For those of us who like to read
and not think, the result is rather straightforward as it will always run the migration, either when forced (because you
can use `oh-my-posh config migrate` manually), or when the current config version doesn't match the required config
version. _Simple, right?_ And this works according to expectations when updating Oh My Posh. The problem starts when you
have a shared config, and use multiple versions of Oh My Posh on the same machine.

Imagine working using multiple shells, one is WSL Ubuntu with a version of Oh My Posh that's still on version 1
(**S1**). One day, you update PowerShell in Windows to use the latest greatest, using version 2 (**S2**). Starting S2,
**it automatically migrates the config to version 2**, moving the `template` property to the segment's model.

In practice, the following version 1 config:

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "version": 1,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "background": "#9A348E",
          "foreground": "#ffffff",
          "leading_diamond": "\ue0b6",
          "properties": {
            // highlight-next-line
            "template": "{{ .UserName }} "
          },
          "style": "diamond",
          "type": "session"
        }
      ]
    }
  ]
}
```

Is migrated to:

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "version": 2,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "background": "#9A348E",
          "foreground": "#ffffff",
          "leading_diamond": "\ue0b6",
          // highlight-next-line
          "template": "{{ .UserName }} ",
          "style": "diamond",
          "type": "session"
        }
      ]
    }
  ]
}
```

Awesome, no user interaction at all and _everything works as designed_. However, you now want to do something using S1, and
due to the logic, Oh My Posh sees that `2 != 1` and **the migration is triggered for version 1**. The migration as such
is non-breaking, but Oh My Posh doesn't know the `template` property on `segment` in this case. The end result of our
config above is rather annoying as we're now left with this:

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "version": 1,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "background": "#9A348E",
          "foreground": "#ffffff",
          "leading_diamond": "\ue0b6",
          "style": "diamond",
          "type": "session"
        }
      ]
    }
  ]
}
```

😱 the template is gone! Nothing to worry about just yet, Oh My Posh will use the default template so you will still see
a prompt, and the migration also _creates a backup_ of your previously correct version 2 config. Up to this point you
can still revert, it's unfortunate, but your local tweaks are still available. If you happen to continue working in S2 at
this point however, it will once again trigger the migration to version 2 and **overwrite your backup with the faulty
version 1 config**.

While the fix for this issue in code is trivial, it's impossible to push that to older versions. People will upgrade
to the most recent version, and in this specific setup, this is potentially breaking. **The advice is to upgrade all versions
at once**, so you only migrate once and there's no "old" version of Oh My Posh left to trigger this unwanted side-effect.

The actual fix for the issue can be found in [7.52.1][fix-version], but due to obvious reasons, it will only work
going from version 2 to any new config version in the future.

```go
if !env.Flags().Migrate && cfg.Version < configVersion {
    cfg.BackupAndMigrate(env)
}
```

While I try hard to come up with every possible edge case when validating this logic, sometimes things slip through.
Even when looking at this one **I find it challenging not having considered that case**, but it still happened. The
good news? It won't happen again 😅

[fix-version]: https://github.com/JanDeDobbeleer/oh-my-posh/releases/tag/v7.52.1
