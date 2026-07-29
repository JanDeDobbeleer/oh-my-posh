<!-- markdownlint-disable-next-line MD041 -->
A [Homebrew][brew] Formula and Cask are available for easy installation.

```bash
brew install jandedobbeleer/oh-my-posh/oh-my-posh
```

Updating is done via:

```bash
brew update && brew upgrade oh-my-posh
```

:::tip
In case you see [strange behaviour][strange] in your shell, reload it after upgrading Oh My Posh.
For example in zsh:

```bash
brew update && brew upgrade && exec zsh
```

:::

[brew]: https://brew.sh
[strange]: https://github.com/JanDeDobbeleer/oh-my-posh/issues/1287
