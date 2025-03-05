---
title: Performance Tips
description: Optimize Oh My Posh for faster prompt rendering across platforms.
---

## Improving Prompt Performance

*Oh My Posh* is designed to render your shell prompt quickly, but performance can vary depending on your operating system, shell, and configuration. This section explores ways to optimize your experience and previews upcoming features to boost speed.

### Understanding Process Overhead

Every time your prompt renders, *Oh My Posh* launches a small executable to generate the output. On Linux and macOS, this process is fast and barely noticeable. On Windows, however, starting a new process can take longer due to how the operating system manages processes—sometimes worsened by security software like antivirus programs.

To see this in action:
1. Open a terminal with *Oh My Posh* enabled (e.g., PowerShell).
2. Hold down the Enter key to spam new prompts.
3. Notice how quickly (or slowly) new lines appear compared to a Linux/macOS system.

For Windows users, this lag can make the prompt feel less responsive, especially in rapid command sequences.

### Proposed Solution: Server Mode (#6124)

A community suggestion (see [issue #6124](https://github.com/JanDeDobbeleer/oh-my-posh/issues/6124)) aims to address this by introducing a "server mode" for *Oh My Posh*. Instead of launching a new process for every prompt, a single, long-running *Oh My Posh* process would stay active for your terminal session. Your shell would send commands to it via a fast communication method (like Unix domain sockets), getting the prompt back without the startup delay.

#### Benefits
- **Windows Users**: Significant speed boost by avoiding repeated process launches.
- **All Platforms**: Slightly faster rendering, even on Linux/macOS.
- **Future-Proofing**: A foundation for more advanced features.

#### How It Might Work
- Launch *Oh My Posh* in "server mode" when you start your shell.
- Use a simple command (e.g., `oh-my-posh server`) to keep it running.
- Your shell talks to it behind the scenes—no extra work for you!

#### Status
This feature is still in discussion. Challenges include making it work across all shells (PowerShell, Bash, CMD, etc.) and ensuring it’s easy to use. We’d love your feedback—try commenting on [#6124](https://github.com/JanDeDobbeleer/oh-my-posh/issues/6124) with your thoughts or experiences!

### Current Tips for Windows Users
While server mode is explored, try these to speed up your prompt:
- **Simplify Your Theme**: Reduce segments (e.g., disable `fetch_status` in the `git` segment).
- **Use a Fast Shell**: PowerShell 7 or Windows Terminal can feel snappier than CMD.
- **Debug Performance**: Run `oh-my-posh debug` to see which segments take the most time.

Stay tuned for updates on #6124 as the community and maintainers work toward a faster *Oh My Posh*!

---
Related issue: #6124