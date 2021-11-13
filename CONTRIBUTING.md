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

## Code of Conduct

### Our Pledge

We as maintainers and contributors pledge to make participation in our
community a harassment-free experience for everyone, regardless of age, body
size, visible or invisible disability, ethnicity, sex characteristics, gender
identity and expression, level of experience, education, socio-economic status,
nationality, personal appearance, race, religion, or sexual identity
and orientation.

We pledge to act and interact in ways that contribute to an open, welcoming,
diverse, inclusive, and healthy community.

### Our Standards

Examples of behavior that contributes to a positive environment for the
project include:

- Demonstrating empathy and kindness toward other people
- Being respectful of differing opinions, viewpoints, and experiences
- Giving and gracefully accepting constructive feedback
- Accepting responsibility and apologizing to those affected by our mistakes,
  and learning from the experience
- Focusing on what is best not just for us as individuals, but for the
  project

Examples of unacceptable behavior include:

- The use of sexualized language or imagery, and sexual attention or
  advances of any kind
- Trolling, insulting or derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information, such as a physical or email
  address, without their explicit permission
- Other conduct which could reasonably be considered inappropriate in a
  professional setting

### Enforcement Responsibilities

Project maintainers are responsible for clarifying and enforcing our standards of
acceptable behavior and will take appropriate and fair corrective action in
response to any behavior that they deem inappropriate, threatening, offensive,
or harmful.

Project maintainers have the right and responsibility to remove, edit, or reject
comments, commits, code, documentation edits, issues, and other contributions that are
not aligned to this Code of Conduct, and will communicate reasons for moderation
decisions when appropriate.

### Scope

This Code of Conduct applies within all project spaces, and also applies when
an individual is officially representing the project in public spaces.
Examples of representing the project include using an official e-mail address,
posting via an official social media account, or acting as an appointed
representative at an online or offline event.

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported to the project maintainers responsible for enforcement via
[email][conduct].
All complaints will be reviewed and investigated promptly and fairly.

All project maintainers are obligated to respect the privacy and security of the
reporter of any incident.

### Enforcement Guidelines

Project maintainers will follow these Project Impact Guidelines in determining
the consequences for any action they deem in violation of this Code of Conduct:

#### 1. Correction

**Project Impact**: Use of inappropriate language or other behavior deemed
unprofessional or unwelcome in the project.

**Consequence**: A private, written warning from project maintainers, providing
clarity around the nature of the violation and an explanation of why the
behavior was inappropriate. A public apology may be requested.

#### 2. Warning

**Project Impact**: A violation through a single incident or series
of actions.

**Consequence**: A warning with consequences for continued behavior. No
interaction with the people involved, including unsolicited interaction with
those enforcing the Code of Conduct, for a specified period of time. This
includes avoiding interactions in project spaces as well as external channels
like social media. Violating these terms may lead to a temporary or
permanent ban.

#### 3. Temporary Ban

**Project Impact**: A serious violation of project standards, including
sustained inappropriate behavior.

**Consequence**: A temporary ban from any sort of interaction or public
communication with the project for a specified period of time. No public or
private interaction with the people involved, including unsolicited interaction
with those enforcing the Code of Conduct, is allowed during this period.
Violating these terms may lead to a permanent ban.

#### 4. Permanent Ban

**Project Impact**: Demonstrating a pattern of violation of project
standards, including sustained inappropriate behavior,  harassment of an
individual, or aggression toward or disparagement of classes of individuals.

**Consequence**: A permanent ban from any sort of public interaction within
the project.

### Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage],
version 2.0, available [here][coc].

Community Impact Guidelines were inspired by [Mozilla's code of conduct
enforcement ladder](https://github.com/mozilla/diversity).

[docs]: https://ohmyposh.dev/docs
[guide]: https://ohmyposh.dev/docs/contributing_started
[cc]: https://www.conventionalcommits.org/en/v1.0.0/#summary
[homepage]: https://www.contributor-covenant.org
[conduct]: mailto:conduct@ohmyposh.dev
[coc]: https://www.contributor-covenant.org/version/2/0/code_of_conduct.html
[codespaces]: https://github.com/features/codespaces
[devcontainer-ext]: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers
[timezones]: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
[themes]: https://ohmyposh.dev/docs/themes
