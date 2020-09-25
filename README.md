# A prompt theme engine for any shell

![example workflow name](https://github.com/jandedobbeleer/go-my-posh/workflows/Release/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jandedobbeleer/go-my-posh)](https://goreportcard.com/report/github.com/jandedobbeleer/go-my-posh)

What started as the offspring of [oh-my-posh][oh-my-posh] for PowerShell resulted in a cross platform, highly customizable and extensible prompt theme engine. After 4 years of working on oh-my-posh, a modern and more efficient tool was needed to suit my personal needs.

## ❤ Support ❤

[![Patreon][patreon-badge]][patreon]
[![Liberapay][liberapay-badge]][liberapay]
[![Ko-Fi][kofi-badge]][kofi]

## Features

* Git status indications
* Failed command indication
* Admin indication
* Current session indications
* Configurable

## Table of Contents**

* [Installation](#installation)
  * [PowerShell](#powershell)
  * [Precompiled Binaries](#precompiled-binaries)
    * [ZSH](#zsh)
    * [Bash](#bash)
    * [Fish](#fish)
    * [Nix](#nix)
* [Configuration](#configuration)
* [Roadmap](#roadmap)
* [Thanks](#thanks)

## Installation

`go-my-posh` uses ANSI color codes under the hood, these should work everywhere,
but you may have to set your $TERM to `xterm-256color` for it to work.

For maximum enjoyment, make sure to install a powerline enabled font. The fonts I use are patched by [Nerd Fonts][nerdfonts], which offers a maximum of icons you can use to configure your prompt.

### Powershell

A PowerShell module is available for your enjoyment. Install and use it with the following commands.

```powershell
Install-Module -Name go-my-posh -Repository PSGallery
Import-Module go-my-posh
Set-PoshPrompt paradox
```

The `Set-PoshPrompt` function has autocompletion to assist in correctly typing the correct theme. It accepts either one of the [pre-configured themes][themes], or a path to a theme of your own.

To see available themes, make use of the `Get-PoshThemes` function. This prints out all themes based on your current location/environment.

### Precompiled Binaries

You can find precompiled binaries for all major OS's underneath the
[releases tab][releases]. Installation instruction for the different shells below assumes `<go-my-posh>` points to the go-my-posh binary and you've downloaded the [`jandedobbeleer` theme][jandedobbeleer] to your `$HOME` directory.

On UNIX systems, make sure the binary is executable before using it.

```bash
chmod +x gmp_executable
```

#### Bash

Add the following to your `.bashrc` (or `.profile` on Mac):

```bash
function _update_ps1() {
    PS1="$(<go-my-posh> -config ~/jandedobeleer.json -error $?)"
}

if [ "$TERM" != "linux" ] && [ -f <go-my-posh> ]; then
    PROMPT_COMMAND="_update_ps1; $PROMPT_COMMAND"
fi
```

#### ZSH

Add the following to your `.zshrc`:

```bash
function powerline_precmd() {
    PS1="$(<go-my-posh> -config ~/jandedobbeleer.json --error $?)"
}

function install_powerline_precmd() {
  for s in "${precmd_functions[@]}"; do
    if [ "$s" = "powerline_precmd" ]; then
      return
    fi
  done
  precmd_functions+=(powerline_precmd)
}

if [ "$TERM" != "linux" ]; then
    install_powerline_precmd
fi
```

#### Fish

Redefine `fish_prompt` in `~/.config/fish/config.fish`:

```bash
function fish_prompt
    eval <go-my-posh> -config ~/jandedobbeleer.json -error $status
end
```

#### Nix

When using `nix-shell --pure`, `go-my-posh` will not be accessible, and
your prompt will not appear.

As a workaround you can add this snippet to your `.bashrc`,
which should re-enable the prompt in most cases:

```bash
# Workaround for nix-shell --pure
if [ "$IN_NIX_SHELL" == "pure" ]; then
    if [ -x <go-my-posh> ]; then
        alias powerline-go="<go-my-posh> -config ~/jandedobbeleer.json"
    fi
fi
```

## Configuration

As the documentation for all the different segments is still lacking, have a look at the available [themes][themes] for reference.

Every segment has its own properties you can set/override. Have a look at the code for any you would want to tweak, available options
are listed as the `Property` constant with their respective `JSON` notation for use in a segment's `properties` section. Additionally,
a few general properties are available cross segments which can be found in `properties.go`.

## Roadmap

* [x] CI
* [x] Github Releases
* [x] Create documentation for manual installation
* [ ] Create documentation on the different segments
* Create easy installation packages
  * [x] Powershell
  * [ ] Brew
  * [ ] Chocolatey

## Thanks

* [Chris Benti][chrisbenti-psconfig] for providing the first influence to start oh-my-posh
* [Keith Dahlby][keithdahlby-poshgit] for creating posh-git and making life more enjoyable
* [Robby Russel][oh-my-zsh] for creating oh-my-zsh, without him this would probably not be here
* [Janne Mareike Koschinski][justjanne] for providing information on how to get certain information using Go (and the amazing [README][powerline-go])

[oh-my-posh]: https://github.com/JanDeDobbeleer/oh-my-posh
[patreon-badge]: https://img.shields.io/badge/Support-Become%20a%20Patreon!-red.svg
[patreon]: https://www.patreon.com/jandedobbeleer
[liberapay-badge]: https://img.shields.io/badge/Liberapay-Donate-%23f6c915.svg
[liberapay]: https://liberapay.com/jandedobbeleer
[kofi-badge]: https://img.shields.io/badge/Ko--fi-Buy%20me%20a%20coffee!-%2346b798.svg
[kofi]: https://ko-fi.com/jandedobbeleer
[releases]: https://github.com/JanDeDobbeleer/go-my-posh/releases
[jandedobbeleer]: https://github.com/JanDeDobbeleer/go-my-posh/blob/master/themes/jandedobbeleer.json
[themes]: https://github.com/JanDeDobbeleer/go-my-posh/tree/master/themes
[chrisbenti-psconfig]: https://github.com/chrisbenti/PS-Config
[keithdahlby-poshgit]: https://github.com/dahlbyk/posh-git
[oh-my-zsh]: https://github.com/robbyrussell/oh-my-zsh
[justjanne]: https://github.com/justjanne
[powerline-go]: https://github.com/justjanne/powerline-go
