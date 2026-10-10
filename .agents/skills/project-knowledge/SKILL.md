---
name: project-knowledge
description: >
  Verified, actionable gotchas for oh-my-posh. Read the matching topic file BEFORE changing shell integration
  scripts (pwsh, zsh, bash, fish, cmd/Clink, nu, elvish, xonsh), terminal or pty behavior, WSL or e2e shell tests,
  Go internals (cache, segments, streaming, serve daemon, template rendering, output sanitization), or release, CI,
  build, and packaging workflows.
---

# Project Knowledge

Non-obvious constraints, platform quirks, traps, and failed approaches. Read only the files the task touches; any
shell-script edit also needs `testing`. Entries are point-in-time: verify against the code and fix drift.

| Topic                                | Read when                                                                     |
| ------------------------------------ | ----------------------------------------------------------------------------- |
| [codebase](references/codebase.md)   | Go internals: templates, sanitization, segments, cache, streaming, serve      |
| [release](references/release.md)     | `release.yml`, signing, packaging, CI gates, `-race`, rebasing on Windows     |
| [testing](references/testing.md)     | Any shell-script edit; e2e/pty harnesses, WSL, ConPTY, Windows Terminal, nu   |
| [zsh](references/zsh.md)             | `omp.zsh`, zle widgets, coproc, zsh-vi-mode                                   |
| [pwsh](references/pwsh.md)           | `omp.ps1`, PSReadLine, runspaces, engine events, pwsh perf                    |
| [fish](references/fish.md)           | `omp.fish`, fish jobs, fifos, key bindings                                    |
| [bash](references/bash.md)           | `omp.bash`, PROMPT_COMMAND, readline, ble.sh                                  |
| [cmd-clink](references/cmd-clink.md) | `omp.lua`, Clink, Windows pipe lifecycles                                     |

## Write-back

- Append only verified entries that change what an agent does, phrased "do X / don't do Y, because Z", to the
  matching file (new file plus index row only when none fits). No narratives or history.
- Update or delete a stale entry instead of stacking a correction. Date it only when staleness matters.
- Commit knowledge updates with the change they relate to.
