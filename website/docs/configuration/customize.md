---
title: Customizing Your Prompt
description: Learn how to override theme defaults with a custom configuration file.
---

## Overriding Theme Defaults

*Oh My Posh* lets you personalize your prompt by combining built-in themes with custom settings. If a theme (e.g., `star.omp.json`) doesn’t fully suit you, you can override its defaults using a custom config file.

### How It Works
1. Themes are loaded from the `themes/` directory or a built-in set.
2. A custom config file (e.g., `~/.myposh.omp.json`) can override specific properties like colors, segments, or styles.
3. When you run `oh-my-posh init` or set `$POSH_THEME`, the custom file takes precedence.

### Steps to Override a Theme
1. **Pick a Base Theme**:
   - Find a theme you like (e.g., `themes/star.omp.json`) in the repo or your local install (`oh-my-posh get themes`).
   - Copy it to a new file:
     ```bash
     cp ~/.poshthemes/star.omp.json ~/.myposh.omp.json
