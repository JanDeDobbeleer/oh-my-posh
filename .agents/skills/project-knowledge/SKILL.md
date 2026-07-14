---
name: project-knowledge
description: >
  Accumulated project memory: verified gotchas and prior findings for oh-my-posh work. Consult
  BEFORE touching shell integration scripts (zsh, pwsh, fish, bash, cmd/Clink), terminal or pty
  behavior, WSL-based shell testing, or internals (cache, segments, streaming, serve daemon).
  Read only the topic files relevant to the task at hand.
---

# Project Knowledge

A modular memory of hard-won, verified facts about this codebase and its runtime environments.
Everything here was learned the expensive way - through debugging, benchmarking, or reproduction -
and is not derivable from a quick read of the code.

## How to use

1. Identify which topics the task touches (a zsh script change touches `zsh` and probably
   `testing`; a segment bug touches `codebase`).
2. Read those reference files before writing code or designing an experiment.
3. Treat entries as point-in-time observations: they carry dates, and code moves. Verify a claim
   against the current code before building on it, and fix the entry when it drifted.

## Topics

| Topic                                | Read when                                                              |
| ------------------------------------ | ---------------------------------------------------------------------- |
| [codebase](references/codebase.md)   | Touching Go code: segments, cache, templates, streaming, serve daemon  |
| [zsh](references/zsh.md)             | Touching `omp.zsh`, zle widgets, coproc, or zsh plugin interop         |
| [pwsh](references/pwsh.md)           | Touching `omp.ps1`, PSReadLine, runspaces, events, or pwsh perf        |
| [fish](references/fish.md)           | Touching `omp.fish`, fish jobs, fifos, or fish event handlers          |
| [bash](references/bash.md)           | Touching `omp.bash`, PROMPT_COMMAND, readline, or bash coprocs         |
| [cmd-clink](references/cmd-clink.md) | Touching `omp.lua`, Clink integration, or Windows pipe lifecycles      |
| [terminal](references/terminal.md)   | Reasoning about ptys, ConPTY, Windows Terminal, or terminal encoding   |
| [testing](references/testing.md)     | Building a harness to drive a shell end-to-end (WSL, zpty, script(1))  |

## How to maintain

This is a living memory - extend it whenever a session ends with knowledge worth keeping:

- Add durable, **verified** facts only: gotchas, platform quirks, measured numbers, failed
  approaches worth not retrying. No speculation, no session-specific state.
- Date non-obvious claims (`verified 2026-07-14`) so future readers can judge staleness.
- One topic per file. Append to the matching reference file; create a new file and index row when
  a fact fits no existing topic.
- Prefer updating or deleting a stale entry over stacking corrections on top of it.
- Keep entries self-contained: name the file, function, or command they apply to.
- This skill is committed - include knowledge updates in the commit of the change they relate to.
