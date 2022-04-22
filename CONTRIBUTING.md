# Contributing

Note we have a code of conduct, please follow it in all your interactions with the project.

Ensure you've read through the [documentation][docs] so you understand the core concepts of the
project. If you're looking to get familiar with go, following the getting started [guide][guide]
can be a good starting point.

## Pull Request Process

1. Ensure any dependencies or build artifacts are removed/ignored before creating a commit.
2. Commits follow the [conventional commits][cc] guidelines.
3. Update the documentation with details of changes to the functionality, this includes new segments
   or core functionality.
4. Pull Requests are merged once all checks pass and a project maintainer has approved it.

## Codespaces / Devcontainer Development Environment

Arguably the easiest way to contribute anything is to use our prepared development environment.

We have a `.devcontainer/devcontainer.json` file, meaning we are compatible with:

- [Github Codespaces][codespaces], or
- the [Visual Studio Code Remote - Containers][devcontainer-ext] extension.

This Linux environment includes all shells supported by oh-my-posh, including Bash, ZSH,
Fish and PowerShell, the latter of which is the default.

### Configuring Devcontainer's Timezone & Theme

1. Open the `.devcontainer/devcontainer.json` file and in the "*build*" section modify:

   - `TZ`: with [your own timezone][timezones]
   - `DEFAULT_POSH_THEME`: with [your preferred theme][themes]

2. Summon the Command Panel (Ctrl+Shift+P) and select `Codespaces: Rebuild Container`
   to rebuild your devcontainer. (This should take just a few seconds.)

### Recompiling oh-my-posh within

The devcontainer definition preinstalls the latest stable oh-my-posh release at build time.

To overwrite the installation's version inside the running devcontainer, you may use the
VSCode *task* `devcontainer: build omp` to rebuild your oh-my-posh with that of
your running repository's state. (You might see a button for this in your statusbar.)

If the compile succeeds, `oh-my-posh --version` should reply:
`development`

Should you somehow mess up your devcontainer's OMP install catastrophically, remember that
if you do `Codespaces: Rebuild Container` again, you'll be back to the latest stable release.

[docs]: https://ohmyposh.dev/docs
[guide]: https://ohmyposh.dev/docs/contributing_started
[cc]: https://www.conventionalcommits.org/en/v1.0.0/#summary
[conduct]: mailto:conduct@ohmyposh.dev
[codespaces]: https://github.com/features/codespaces
[devcontainer-ext]: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers
[timezones]: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
[themes]: https://ohmyposh.dev/docs/themes
