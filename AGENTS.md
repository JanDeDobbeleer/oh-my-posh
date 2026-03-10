# Agent Instructions

## APM Setup

This repository uses [APM](https://github.com/JanDeDobbeleer/agentic) to manage agent skills.
Before starting any task, verify that the skills listed in `apm.yml` are installed under
the `apm_modules/` directory.

If `apm_modules/` is missing or any skill package from `apm.yml` is not present, install them by running:

```sh
pip install apm-cli
apm install
```

## General File Creation Guidelines

When creating new files:

- **Always use LF (Unix-style) line endings**, not CRLF (Windows-style)
- This repository uses `.gitattributes` to enforce LF line endings
- Ensures consistency across all platforms and avoids Git warnings
