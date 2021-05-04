---
id: introduction
title: Introduction
sidebar_label: 👋 Introduction
slug: /
---

Oh my Posh is a custom prompt engine for any shell that has the ability to adjust
the prompt string with a function or variable.

## Oh my Posh 3 vs Oh my Posh 2

Oh my Posh 3 is the offspring of [Oh my Posh][omp], a prompt theme engine for PowerShell.
Oh my Posh started out by being inspired from tools like [oh my zsh][omz] when nothing was
available specifically for PowerShell.

Over the years, I switched operating system/main shell quite a lot, even on
Windows via the [WSL][wsl]. This made it so that my prompt wasn't portable enough,
I wanted the same visual/functional experience regardless
of the shell I was working in. Hello world [Oh my Posh 3][omp3]!

## Concept

Traditionally, prompt tools work with custom scripts per theme (just like [Oh my Posh][omp] did) or a lot
of CLI configuration switches to define what it looks like. With Oh my Posh, I wanted to start from a single
configuration file that could easily be shared anywhere, removing the need to really grasp what goes on underneath.

When you look at prompts like Agnoster or Paradox, you notice they usually consist of a few building
**blocks** which contain one or more **segments** that display some sort of information. The configuration of
Oh my Posh works exactly like that. Blocks are a combination of one or more segments.

The basic layout of the config file is as follows.

```json
{
    "blocks": []
}
```

A [block][block] has properties that indicate its position and the [segments][segment] it will render.

```json
{
    "blocks": [
        {
            // positioning metadata (not shown)
            "segments": []
        }
    ]
}
```

A [segment][segment] renders a single context like showing the current folder, user information or git status
when relevant. It can be styled any way you want, resulting in visualizing the prompt you are looking for.

For your convenience, the existing [themes][themes] from [Oh my Posh][omp-themes] have been added to version 3, so you
can get started even without having to understand the theming. So, let's no longer waste time on theory, have a look at the
[installation guide][install] to get started right away!

[omp]: https://github.com/JanDeDobbeleer/oh-my-posh2
[omz]: https://github.com/ohmyzsh/ohmyzsh
[omp3]: https://github.com/JanDeDobbeleer/oh-my-posh
[wsl]: https://docs.microsoft.com/en-us/windows/wsl/install-win10
[install]: /docs/installation
[block]: /docs/configure#block
[segment]: /docs/configure#segment
[themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes
[omp-themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/master/Themes
