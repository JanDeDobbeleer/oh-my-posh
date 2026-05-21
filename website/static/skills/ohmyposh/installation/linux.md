# Oh My Posh — Linux Installation

> **Prerequisites:** `curl`, `unzip`, `realpath`, and `dirname` must be available.
> Make sure `curl` and the client certificate store are up to date.

## Install

```bash
curl -s https://ohmyposh.dev/install.sh | bash -s
```

By default this installs to `~/bin` or `~/.local/bin` (whichever exists), or to the directory
of any existing `oh-my-posh` binary. To install to a custom location:

```bash
curl -s https://ohmyposh.dev/install.sh | bash -s -- -d ~/bin
```

## Update

Re-run the install script — it replaces the existing binary in-place.

## After installing

Restart the terminal (or open a new window) so `oh-my-posh` is on `$PATH`, then proceed to the
[shell setup guide](/skills/ohmyposh/SKILL.md).
