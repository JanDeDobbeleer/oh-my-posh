# API Data Directory

This directory contains data files used by the Azure Functions API.

## Schema File

The `schema.json` file is automatically copied from `themes/schema.json` during the GitHub Actions
deployment workflow. This file is used by the MCP validator to validate oh-my-posh configurations.

The schema is embedded locally to:

- Improve performance (no external HTTP requests)
- Ensure reliability (no dependency on external services)
- Work offline/in isolated environments
