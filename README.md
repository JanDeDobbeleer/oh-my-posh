# A prompt theme engine for any shell

![Release Status][release-status]
[![Go Report Card][report-card]][report-card-link]
[![PS Gallery][psgallery-badge]][powershell-gallery]
[![Documentation][docs-badge]][docs]

What started as the offspring of [oh-my-posh2][oh-my-posh2] for PowerShell resulted in a cross platform,
highly customizable and extensible prompt theme engine. After 4 years of working on oh-my-posh,
a modern and more efficient tool was needed to suit my personal needs.

## ❤ Support ❤

[![Patreon][patreon-badge]][patreon]
[![Liberapay][liberapay-badge]][liberapay]
[![Ko-Fi][kofi-badge]][kofi]

## Features

* Shell independent
* Git status indications
* Failed command indication
* Admin indication
* Current session indications
* Language info
* Shell info
* Configurable

## Documentation

[![Documentation][docs-badge]][docs]

## Roadmap

* [x] CI
* [x] Github Releases
* [x] Create documentation for manual installation
* [x] Create documentation on the different segments
* Create easy installation packages
  * [x] Powershell
  * [x] Brew
  * [x] Scoop
  * [ ] Winget
* [x] Swap V2 with V3

## Thanks

* [Chris Benti][chrisbenti-psconfig] for providing the first influence to start oh-my-posh
* [Keith Dahlby][keithdahlby-poshgit] for creating posh-git and making life more enjoyable
* [Robby Russel][oh-my-zsh] for creating oh-my-zsh, without him this would probably not be here
* [Janne Mareike Koschinski][justjanne] for providing information on how to get certain information
using Go (and the amazing [README][powerline-go])
* [Starship][starship] for creating an amazing way to initialize the prompt

[release-status]: https://github.com/jandedobbeleer/oh-my-posh/workflows/Release/badge.svg
[psgallery-badge]: https://img.shields.io/powershellgallery/dt/oh-my-posh.svg
[powershell-gallery]: https://www.powershellgallery.com/packages/oh-my-posh/
[report-card]: https://goreportcard.com/badge/github.com/jandedobbeleer/oh-my-posh
[report-card-link]: https://goreportcard.com/report/github.com/jandedobbeleer/oh-my-posh
[oh-my-posh2]: https://github.com/JanDeDobbeleer/oh-my-posh2
[patreon-badge]: https://img.shields.io/badge/Support-Become%20a%20Patreon!-red.svg
[patreon]: https://www.patreon.com/jandedobbeleer
[liberapay-badge]: https://img.shields.io/badge/Liberapay-Donate-%23f6c915.svg
[liberapay]: https://liberapay.com/jandedobbeleer
[kofi-badge]: https://img.shields.io/badge/Ko--fi-Buy%20me%20a%20coffee!-%2346b798.svg
[kofi]: https://ko-fi.com/jandedobbeleer
[docs-badge]: https://img.shields.io/badge/documentation-ohmyposh.dev-blue
[docs]: https://ohmyposh.dev/docs
[chrisbenti-psconfig]: https://github.com/chrisbenti/PS-Config
[keithdahlby-poshgit]: https://github.com/dahlbyk/posh-git
[oh-my-zsh]: https://github.com/robbyrussell/oh-my-zsh
[justjanne]: https://github.com/justjanne
[powerline-go]: https://github.com/justjanne/powerline-go
[starship]: https://github.com/starship/starship/blob/master/src/init/mod.rs
