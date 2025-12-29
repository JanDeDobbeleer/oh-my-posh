---
date: 2025-12-28
description: Oh My Posh now integrates with Claude Code's statusline feature to bring beautiful, customizable AI session information right into your terminal prompt.
tags:
- claude
- ai
- terminal
- prompt
- statusline
- integration
- customization
title: "Oh My Posh Meets Claude Code: AI-Powered Terminal Prompts"
slug: oh-my-posh-claude-code-integration
authors:
- name: Jan De Dobbeleer
  title: Maintainer
  url: https://github.com/jandedobbeleer
  image_url: https://avatars.githubusercontent.com/u/2492783?v=4
---

Terminal customization just got a lot smarter. Oh My Posh now integrates with [Claude Code](https://code.claude.com/)
through its [statusline functionality](https://code.claude.com/docs/en/statusline), bringing real-time AI session
information  and development context directly into your Claude Code prompt.
This isn't just another segment: it's a bridge between your
development workflow and AI-powered coding assistance.

![Claude Code](/img/claude.png)

## What is Claude Code's statusline?

Claude Code's statusline feature allows you to create custom status displays that appear at the bottom of
the Claude Code interface, similar to how terminal prompts work in shells. The statusline receives rich JSON
data about your current AI session via stdin, including:

- **Model information**: Which Claude model you're using (Claude Sonnet, Claude Opus, etc.)
- **Token usage**: Input/output tokens, context window utilization, and usage percentages
- **Cost tracking**: Real-time cost calculations and session duration
- **Workspace context**: Current and project directories
- **Session metadata**: Unique session IDs and version information

The statusline updates automatically when conversation messages change (throttled to every 300ms max), and
your command's stdout becomes the status display with full ANSI color support.

### Setting Up the Integration

Configuration is straightforward. Add this to your Claude Code settings:

```json title="~/.claude/settings.json"
{
  "statusLine": {
    "type": "command",
    "command": "oh-my-posh claude",
    "padding": 0
  }
}
```

That's it! Oh My Posh will automatically detect when Claude Code provides session data and display the
relevant information in your prompt.

It's important to note that the `claude` CLI command operates differently from your regular prompt
configuration. When used as a statusline command, Oh My Posh runs in a special mode that's completely separate
from your standard terminal prompt. This means you'll likely want to create a dedicated, minimal configuration
specifically for Claude Code that focuses on displaying AI session information rather than your usual prompt
elements.

By default, the `oh-my-posh claude` command provides a built-in statusline that shows your current working
directory, git context, active model name, and context window usage as a visual gauge. To customize this
display, use the `--config` flag to specify your own theme configuration file that includes a custom claude
segment tailored to your preferences.

## The Claude Code segment

Oh My Posh's new `claude` segment taps into this statusline data to bring AI session awareness directly
into your terminal prompt. When you use the `oh-my-posh claude` command as your statusline command in Claude
Code, you get access to a wealth of session information that can be displayed in your prompt without needing
to know the technical details.

### Example Configuration

Here's a sample configuration that shows the model name and context usage:

```json
{
  "type": "claude",
  "style": "diamond",
  "leading_diamond": "\ue0b6",
  "trailing_diamond": "\ue0b4",
  "foreground": "#FFFFFF",
  "background": "#FF6B35",
  "template": " \udb82\udfc9 {{ .Model.DisplayName }} \uf2d0 {{ .TokenUsagePercent.Gauge }} "
}
```

This displays something like: ` 🤖 Claude 4.5 Sonnet  ▰▰▰▱▱ `

The gauge provides instant visual feedback on how much of your context window you've consumed, which is
crucial for managing long coding sessions.

## The Technical Details

Under the hood, Oh My Posh reads the rich JSON session data that Claude Code provides via stdin when used
as a statusline command. The segment only activates when Claude Code session data is available, so there's no
performance impact when you're not using Claude Code.

## Getting Started

If you're already using Oh My Posh, adding Claude Code integration is as simple as:

1. Install Claude Code if you haven't already
2. Add the statusline configuration to your Claude Code settings
3. Optionally create your own configuration including the `claude` segment
4. Start a Claude Code session and watch your prompt come alive

For detailed configuration options and all available properties, check out the [complete Claude segment documentation](https://ohmyposh.dev/docs/segments/cli/claude).

## What's Next?

This integration opens up exciting possibilities. Imagine prompts that:

- Change color based on token usage percentage
- Show different icons for different AI models
- Display cost warnings when sessions get expensive
- Integrate with any other segment to show additional development context

The foundation is there, and now it's up to the community to build amazing configurations that make
AI-powered development even more seamless.
