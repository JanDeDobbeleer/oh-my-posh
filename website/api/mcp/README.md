# MCP Validator Function

This directory contains the Azure Function that implements the Model Context Protocol (MCP) server for validating oh-my-posh configurations.

## Endpoints

- `POST /api/mcp` - MCP server endpoint that handles validation requests
- `GET /api/mcp` - Returns server information and available tools

## Supported Tools

1. **validate_config** - Validate an oh-my-posh configuration
   - Supports JSON, YAML, and TOML formats
   - Returns detailed validation errors with JSON paths

## Usage

### As an MCP Server

Configure your MCP client to connect to this server:

```json
{
  "mcpServers": {
    "oh-my-posh-validator": {
      "url": "https://ohmyposh.dev/api/mcp",
      "transport": "http"
    }
  }
}
```

### Direct HTTP Requests

#### Get Server Info

```bash
curl https://ohmyposh.dev/api/mcp
```

#### List Available Tools

```bash
curl -X POST https://ohmyposh.dev/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 1
  }'
```

#### Validate a Configuration

```bash
curl -X POST https://ohmyposh.dev/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "validate_config",
      "arguments": {
        "content": "{\"$schema\":\"https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json\",\"blocks\":[]}",
        "format": "json"
      }
    },
    "id": 1
  }'
```

## Response Format

The validation result includes:

- `valid`: Boolean indicating if the configuration is valid
- `errors`: Array of validation errors (if any)
- `warnings`: Array of warnings (best practices, deprecations)
- `detectedFormat`: The detected or specified format
- `parsedConfig`: The parsed configuration object (for debugging)

Example response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{
          \"valid\": true,
          \"errors\": [],
          \"warnings\": [
            {
              \"path\": \"$schema\",
              \"message\": \"Consider adding \\\"$schema\\\" property for better editor support.\",
              \"type\": \"recommendation\"
            }
          ],
          \"detectedFormat\": \"json\",
          \"parsedConfig\": {...}
        }"
      }
    ]
  },
  "id": 1
}
```

## Development

To test locally:

```bash
cd website/api
npm install
npm start
```

Then send requests to `http://localhost:7071/api/mcp`
