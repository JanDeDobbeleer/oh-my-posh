---
id: dsc
title: Desired State Configuration
sidebar_label: 🖥️ Desired State Configuration
---

Oh My Posh supports Desired State Configuration (DSC) for declarative configuration management, enabling automated
deployment and consistent configuration across multiple systems.

## Concept

Oh My Posh DSC builds on the traditional Oh My Posh configuration approach by adding automation and orchestration
capabilities. Instead of manually configuring your prompt, you can define the desired state declaratively and let DSC
ensure your system matches that state.

DSC works with **resources** that represent different aspects of your Oh My Posh setup:

- **Configuration Resource**: Manages your Oh My Posh configuration files
- **Shell Resource**: Handles shell initialization and integration
- **Font Resource**: Tracks installed Nerd Fonts

These resources can be used standalone through the CLI or integrated with orchestration tools like WinGet and
Microsoft DSC for automated deployments.

## Overview

DSC support in Oh My Posh provides:

- **Declarative configuration**: Define the desired state of your Oh My Posh setup
- **Automated deployment**: Configure Oh My Posh as part of system provisioning workflows
- **Shell integration**: Automatically configure shell initialization for bash, zsh, fish, PowerShell, and more
- **Font management**: Track installed Nerd Fonts through DSC state
- **Orchestration support**: Integration with WinGet and Microsoft DSC tools

DSC functionality is available through the `oh-my-posh` CLI and can be used standalone or with orchestration tools.

## DSC Resources

Oh My Posh provides the following DSC resources:

### Configuration Resource

Manages Oh My Posh configuration files.

**Operations**: `get`, `set`, `export`, `schema`

```bash
# Get current configuration state
oh-my-posh config dsc get

# Apply a configuration
oh-my-posh config dsc set --state '{"states":[{"source":"~/mytheme.omp.json","format":"json"}]}'

# Get configuration schema
oh-my-posh config dsc schema
```

### Shell Resource

Manages shell initialization and integration.

**Operations**: `get`, `set`, `export`, `schema`

```bash
# Get current shell configurations
oh-my-posh init bash dsc get

# Configure shell initialization
oh-my-posh init bash dsc set --state '{"states":[{"name":"bash","command":"oh-my-posh init bash --config ~/mytheme.omp.json"}]}'
```

### Font Resource

Tracks Nerd Fonts installed through Oh My Posh.

**Operations**: `get`, `export`, `schema`

```bash
# Get installed fonts
oh-my-posh font dsc get

# Get font schema
oh-my-posh font dsc schema
```

## Direct CLI Usage

You can use the DSC commands directly from the command line for configuration management.

### Managing Configurations

#### Get State

Retrieve the current configuration state:

```bash
oh-my-posh config dsc get
```

Example output:

```json
{
  "states": [
    {
      "format": "json",
      "source": "~/mytheme.omp.json"
    }
  ]
}
```

#### Set State

Apply a new configuration state:

```bash
oh-my-posh config dsc set --state '{"states":[{"source":"~/mytheme.omp.json","format":"json"}]}'
```

This creates or updates the configuration file at the specified location with the provided format.

#### Schema

Get the JSON schema for the configuration resource:

```bash
oh-my-posh config dsc schema
```

Use this to understand the structure and available options for configuration states.

### Managing shell Integration

#### Bash

Configure Oh My Posh initialization for bash:

```bash
# Get current state
oh-my-posh init bash dsc get

# Set initialization
oh-my-posh init bash dsc set --state '{"states":[{"name":"bash","command":"oh-my-posh init bash --config ~/mytheme.omp.json"}]}'
```

This automatically updates your `.bashrc` or `.bash_profile` with the Oh My Posh initialization command.

#### Zsh

Configure Oh My Posh initialization for zsh:

```bash
# Get current state
oh-my-posh init zsh dsc get

# Set initialization
oh-my-posh init zsh dsc set --state '{"states":[{"name":"zsh","command":"oh-my-posh init zsh --config ~/mytheme.omp.json"}]}'
```

This automatically updates your `.zshrc` with the Oh My Posh initialization command.

#### PowerShell

Configure Oh My Posh initialization for PowerShell:

```powershell
# Get current state
oh-my-posh init pwsh dsc get

# Set initialization
oh-my-posh init pwsh dsc set --state '{"states":[{"name":"pwsh","command":"oh-my-posh init pwsh --config ~/mytheme.omp.json"}]}'
```

This automatically updates your PowerShell profile with the Oh My Posh initialization command.

#### Fish

Configure Oh My Posh initialization for fish:

```bash
# Get current state
oh-my-posh init fish dsc get

# Set initialization
oh-my-posh init fish dsc set --state '{"states":[{"name":"fish","command":"oh-my-posh init fish --config ~/mytheme.omp.json"}]}'
```

This automatically updates your fish `config.fish` with the Oh My Posh initialization command.

## Orchestration with WinGet

WinGet configuration enables you to install Oh My Posh and apply configuration in a single declarative file.

### Basic WinGet configuration

Create a configuration file to install and configure Oh My Posh:

```yaml title="oh-my-posh-setup.yaml"
$schema: https://raw.githubusercontent.com/PowerShell/DSC/main/schemas/v3/config/document.json
metadata:
  winget:
    processor: dscv3
resources:
  - name: Install Oh My Posh
    type: Microsoft.WinGet.DSC/WinGetPackage
    properties:
      id: JanDeDobbeleer.OhMyPosh
      source: winget
```

Apply the configuration:

```powershell
winget configure oh-my-posh-setup.yaml
```

### Complete setup with configuration and shell

This example installs Oh My Posh, adds your configuration, and initializes PowerShell:

```yaml title="oh-my-posh-complete.yaml"
$schema: https://raw.githubusercontent.com/PowerShell/DSC/main/schemas/v3/config/document.json
metadata:
  winget:
    processor: dscv3
resources:
  - name: Install Oh My Posh
    type: Microsoft.WinGet.DSC/WinGetPackage
    properties:
      id: JanDeDobbeleer.OhMyPosh
      source: winget

  - name: Add Oh My Posh configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/mytheme.omp.json
          format: json

  - name: Initialize PowerShell
    type: OhMyPosh/Shell
    properties:
      states:
        - name: pwsh
          command: oh-my-posh init pwsh --config ~/mytheme.omp.json
```

Apply with:

```powershell
winget configure oh-my-posh-complete.yaml
```

### Multi-shell configuration

Initialize multiple shells with different configurations:

```yaml title="oh-my-posh-multi-shell.yaml"
$schema: https://raw.githubusercontent.com/PowerShell/DSC/main/schemas/v3/config/document.json
metadata:
  winget:
    processor: dscv3
resources:
  - name: Install Oh My Posh
    type: Microsoft.WinGet.DSC/WinGetPackage
    properties:
      id: JanDeDobbeleer.OhMyPosh
      source: winget

  - name: Add work configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/work-theme.omp.json
          format: json

  - name: Add personal configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/personal-theme.omp.json
          format: json

  - name: Initialize PowerShell with work configuration
    type: OhMyPosh/Shell
    properties:
      states:
        - name: pwsh
          command: oh-my-posh init pwsh --config ~/work-theme.omp.json

  - name: Initialize Bash with personal configuration
    type: OhMyPosh/Shell
    properties:
      states:
        - name: bash
          command: oh-my-posh init bash --config ~/personal-theme.omp.json
```

## Orchestration with Microsoft DSC

Microsoft DSC (`dsc`) provides cross-platform configuration management capabilities. Oh My Posh provides native DSC
resources that can be used in DSC configuration documents.

### Example DSC configuration

Create a configuration document for Oh My Posh:

```yaml title="oh-my-posh-dsc.yaml"
$schema: https://aka.ms/dsc/schemas/v3/bundled/config/document.json
resources:
  - name: Add Oh My Posh configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/mytheme.omp.json
          format: json

  - name: Initialize PowerShell
    type: OhMyPosh/Shell
    properties:
      states:
        - name: pwsh
          command: oh-my-posh init pwsh --config ~/mytheme.omp.json
```

Apply the configuration using the `dsc` CLI:

```bash
dsc config set --document oh-my-posh-dsc.yaml
```

### Complete configuration with multiple shells

```yaml title="oh-my-posh-complete-dsc.yaml"
$schema: https://aka.ms/dsc/schemas/v3/bundled/config/document.json
resources:
  - name: Add primary configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/primary-theme.omp.json
          format: json

  - name: Add secondary configuration
    type: OhMyPosh/Config
    properties:
      states:
        - source: ~/secondary-theme.omp.json
          format: yaml

  - name: Initialize PowerShell
    type: OhMyPosh/Shell
    properties:
      states:
        - name: pwsh
          command: oh-my-posh init pwsh --config ~/primary-theme.omp.json

  - name: Initialize Bash
    type: OhMyPosh/Shell
    properties:
      states:
        - name: bash
          command: oh-my-posh init bash --config ~/primary-theme.omp.json

  - name: Initialize Zsh
    type: OhMyPosh/Shell
    properties:
      states:
        - name: zsh
          command: oh-my-posh init zsh --config ~/secondary-theme.omp.json
```

### Resource Types

Oh My Posh provides the following DSC resource types:

#### OhMyPosh/Config

Manages Oh My Posh configuration files.

**Properties:**

- `states` (array): List of configuration states
  - `source` (string): Path to the configuration file
  - `format` (string): Format of the configuration file (`json`, `yaml`, `toml`)

#### OhMyPosh/Shell

Manages shell initialization.

**Properties:**

- `states` (array): List of shell configurations
  - `name` (string): Shell name (`bash`, `zsh`, `pwsh`, `fish`, etc.)
  - `command` (string): Oh My Posh initialization command

#### OhMyPosh/Font

Tracks installed Nerd Fonts. This resource is read-only and automatically populated when fonts are installed through
Oh My Posh.

## Configuration State Management

DSC state is stored in the Oh My Posh cache and persists across sessions. This enables:

- **State tracking**: Oh My Posh remembers configurations set through DSC
- **Idempotency**: Running the same DSC command multiple times produces the same result
- **State validation**: Query current state before making changes

## Advanced Usage

### Multiple configurations

You can manage multiple configuration files:

```bash
oh-my-posh config dsc set --state '{
  "states": [
    {"source":"~/work.omp.json","format":"json"},
    {"source":"~/personal.omp.json","format":"json"}
  ]
}'
```

### Shell-Specific Initialization

Initialize multiple shells with different configuration:

```bash
# Bash with one configuration
oh-my-posh init bash dsc set --state '{"states":[{"name":"bash","command":"oh-my-posh init bash --config ~/bash-theme.omp.json"}]}'

# PowerShell with another configuration
oh-my-posh init pwsh dsc set --state '{"states":[{"name":"pwsh","command":"oh-my-posh init pwsh --config ~/pwsh-theme.omp.json"}]}'
```

## Supported shells

DSC shell configuration supports the following shells:

- **bash**: Configures `.bashrc` or `.bash_profile`
- **zsh**: Configures `.zshrc`
- **fish**: Configures `~/.config/fish/config.fish`
- **pwsh**: Configures PowerShell profile (`$PROFILE`)
- **nu**: Configures `~/.config/nushell/config.nu`
- **elvish**: Configures `.elvish/rc.elv`
- **xonsh**: Configures `.xonshrc`

The shell integration automatically:

- Creates configuration files if they don't exist
- Updates existing Oh My Posh initialization commands
- Preserves shell-specific command syntax
- Maintains proper whitespace and formatting

## See Also

- [General configuration](/docs/configuration/general) - Main configuration documentation
- [Installation](/docs/installation/windows) - Installing Oh My Posh
- [Themes](https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes) - Available themes
- [WinGet configuration](https://learn.microsoft.com/windows/package-manager/configuration/) - WinGet DSC documentation
- [Microsoft DSC](https://learn.microsoft.com/powershell/dsc/overview) - Microsoft DSC documentation
