---
date: 2026-08-03
description: Studio now hands off to the Configurator, git status runs natively for a 10x speedup, and steady dependency cuts shrank the binary by nearly a third.
tags:
- studio
- configurator
- git
- performance
- streaming
- binary-size
title: "Studio, Native Git, and a Leaner Binary"
slug: studio-native-git-and-a-leaner-binary
authors:
- name: Jan De Dobbeleer
  title: Maintainer
  url: https://github.com/jandedobbeleer
  image_url: https://avatars.githubusercontent.com/u/2492783?v=4
---

A few changes landed recently that make Oh My Posh faster to run and easier to configure.
Here's what's new: a live theme editor with a bridge to the Configurator, a native git
integration, a background daemon that keeps your prompt warm, and a binary that keeps getting
smaller.

<!--truncate-->

## Studio: edit your theme, live, in the browser

[`/docs/studio`](/docs/studio) now runs an actual build of Oh My Posh directly in your browser
through WebAssembly. Change your config and the rendered prompt updates immediately. It renders
against recorded sample data to make rendering every segment possible. There's no installation, no
local config file, and no shell restart needed to see what a change looks like.

The editor understands your config the way an IDE understands code. It offers schema-aware
autocomplete for every segment option, in both JSON and YAML, with inline documentation as you
type. Config editing changes from "look up the docs, guess the property name, reload the prompt"
into something closer to writing code with a linter looking over your shoulder.

If you're looking for a more drag and drop experience, **Open in Configurator** hands your config
straight off to the full [Configurator app](https://configurator.ohmyposh.dev/).
Shout out to [James Montemagno] for maintaining that awesome tool!

Studio and the Configurator aren't a linear path where one replaces the other. They're two
different ways of working, and now you can move between them freely. If you want a fast,
code-first way to tweak a theme, Studio's live text editing gets you there fastest. If you want
a more visual experience, dragging segments around and arranging blocks instead of typing, the
Configurator fits better.

## Shrinking the binary, release after release

A few versions ago, the Windows binary was a 19.6 MB download. The latest release is 14.0 MB, a
28% drop. That didn't come from one big rewrite. It came from steadily replacing general-purpose
third-party libraries with small, purpose-built internal code wherever the project only used a
sliver of what a dependency offered. _This isn't a verdict on those libraries_: they're solid,
popular projects. It's about matching the code shipped to the code actually used, since every
user downloads them and loads all of it on every run.

Here's what that looked like in practice, measured on the same stripped linux/amd64 build each
time:

| Dependency replaced | What it did here | Binary size saved |
| --- | --- | --- |
| `hashicorp/hcl/v2` + `zclconf/go-cty` | Parsed a Terraform block to read one attribute | -1.80 MB |
| `bubbletea` + `bubbles` + `lipgloss` + `go-runewidth` | Font picker UI and Unicode width lookup | -1.32 MB (-7.8%) |
| `spf13/cobra` + `spf13/pflag` | Command parsing and flag handling for the CLI | -586 KB |
| `ini.v1`, `x/mod`, `jsonutil`, `google/uuid` | Config reads, a `go.mod` field, comments, ID generation | -266 KB |

Each of these libraries did its job well. It's just that they
were general-purpose tools solving problems narrow enough to write directly, without carrying
their surrounding framework along for the ride. A scanner that reads one field doesn't need a
full HCL parser, and a numbered list doesn't need a terminal UI framework (even when it looks pretty).

Every pull request now gets an automated report comparing its release build against the latest
published binary, per OS. The next dependency that quietly adds a few hundred kilobytes gets
caught before it ships instead of accumulating unnoticed release after release.

## Native git status: what building our own integration buys you

The git segment traditionally works by spawning `git status --porcelain=2` and parsing its
output. On Windows in particular, just spawning that process costs 22-25ms before git has even
looked at a single file: roughly 70% of the segment's total render time.

The new `native_status` option skips the subprocess entirely. It reads the index, refs,
gitignore rules, and object store directly, in-process, and computes the same status counts git
would. Benchmarked head-to-head against the exec path, on real repositories:

| Platform | Repository | Native | `git status` (exec) |
| --- | --- | --- | --- |
| Windows | 1.1k files | 3.4 ms | 38 ms |
| Windows | 12.5k files, 2k ignored | 12 ms | 47 ms |
| Linux (ext4) | 1k files | 1.0 ms | 4.0 ms |
| Linux (ext4) | 10k files, packed | 10 ms | 21 ms |

That's roughly a 10x speedup on Windows and a 2-4x speedup on Linux, on both small and large
repositories. In practice, it's the difference between a git segment that's instant and one you
can feel lag behind your typing, without changing anything about how you use git.

## A background daemon that keeps your prompt warm

The other big cost on every prompt render isn't rendering itself: it's the process start. Every
prompt has, until now, meant launching a fresh `oh-my-posh` process from scratch. Streaming mode
attacks that cost directly.

When you enable `streaming`, your shell starts a single background `oh-my-posh serve` process
once, when the session opens, not once per prompt. That process stays alive for the life of your
shell. It accepts render requests over a lightweight local channel and answers them without
ever paying process-startup cost again. Segment caches stay warm in memory between prompts
instead of being rebuilt from scratch, so the second prompt in a session is often faster than the
first. On top of that, the prompt itself streams: segments that resolve quickly show
immediately, and slower segments show a placeholder and repaint in place the moment their data
lands, instead of the whole prompt waiting on the slowest segment.

Support today spans `powershell`, `zsh`, `fish`, and `cmd` (through Clink), each wired into that
shell's own idioms for talking to a background process. Full docs are up at
[Streaming][streaming].

## Why this matters

Together, these changes make the prompt itself lighter to download, faster to render, and easier
to configure, with the tooling in place, like the binary size report on every PR, to keep it that
way going forward.

[streaming]: /docs/configuration/streaming
[James Montemagno]: https://montemagno.com
