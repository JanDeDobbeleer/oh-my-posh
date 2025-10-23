const validator = require('../shared/validator.js');

/**
 * Azure Function entry point for MCP server
 */
module.exports = async function (context, req) {
  context.log('MCP validator function processed a request');

  // Handle GET requests - return server info
  if (req.method === 'GET') {
    context.res = {
      status: 200,
      headers: {
        'Content-Type': 'application/json'
      },
      body: {
        name: 'oh-my-posh-validator',
        version: '1.0.0',
        description: 'MCP server for validating oh-my-posh configurations',
        capabilities: {
          tools: {}
        },
        tools: [
          {
            name: 'validate_config',
            description: 'Validate an oh-my-posh configuration against the schema',
            inputSchema: {
              type: 'object',
              properties: {
                content: {
                  type: 'string',
                  description: 'The configuration content as a string (JSON, YAML, or TOML)'
                },
                format: {
                  type: 'string',
                  enum: ['json', 'yaml', 'toml', 'auto'],
                  description: 'The format of the configuration (auto-detect if not specified)',
                  default: 'auto'
                }
              },
              required: ['content']
            }
          }
        ]
      }
    };
    return;
  }

  // Handle POST requests - process MCP protocol messages
  try {
    const message = req.body;

    if (!message || !message.jsonrpc || message.jsonrpc !== '2.0') {
      context.res = {
        status: 400,
        body: {
          jsonrpc: '2.0',
          error: {
            code: -32600,
            message: 'Invalid Request: Not a valid JSON-RPC 2.0 message'
          },
          id: message?.id || null
        }
      };
      return;
    }

    // Handle list tools request
    if (message.method === 'tools/list') {
      context.res = {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        },
        body: {
          jsonrpc: '2.0',
          result: {
            tools: [
              {
                name: 'validate_config',
                description: 'Validate an oh-my-posh configuration against the schema. Supports JSON, YAML, and TOML formats.',
                inputSchema: {
                  type: 'object',
                  properties: {
                    content: {
                      type: 'string',
                      description: 'The configuration content as a string (JSON, YAML, or TOML)'
                    },
                    format: {
                      type: 'string',
                      enum: ['json', 'yaml', 'toml', 'auto'],
                      description: 'The format of the configuration (auto-detect if not specified)',
                      default: 'auto'
                    }
                  },
                  required: ['content']
                }
              }
            ]
          },
          id: message.id
        }
      };
      return;
    }

    // Handle tool call request
    if (message.method === 'tools/call') {
      const { name, arguments: args } = message.params;

      if (name !== 'validate_config') {
        context.res = {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          },
          body: {
            jsonrpc: '2.0',
            error: {
              code: -32601,
              message: `Unknown tool: ${name}`
            },
            id: message.id
          }
        };
        return;
      }

      // Validate the configuration
      const result = await validator.validateConfig(args.content, args.format || 'auto');

      context.res = {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        },
        body: {
          jsonrpc: '2.0',
          result: {
            content: [
              {
                type: 'text',
                text: JSON.stringify(result, null, 2)
              }
            ]
          },
          id: message.id
        }
      };
      return;
    }

    // Handle initialize request
    if (message.method === 'initialize') {
      context.res = {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        },
        body: {
          jsonrpc: '2.0',
          result: {
            protocolVersion: '2024-11-05',
            capabilities: {
              tools: {}
            },
            serverInfo: {
              name: 'oh-my-posh-validator',
              version: '1.0.0'
            }
          },
          id: message.id
        }
      };
      return;
    }

    // Unknown method
    context.res = {
      status: 200,
      headers: {
        'Content-Type': 'application/json'
      },
      body: {
        jsonrpc: '2.0',
        error: {
          code: -32601,
          message: `Method not found: ${message.method}`
        },
        id: message.id
      }
    };

  } catch (error) {
    context.log.error('Error processing MCP request:', error);
    context.res = {
      status: 500,
      headers: {
        'Content-Type': 'application/json'
      },
      body: {
        jsonrpc: '2.0',
        error: {
          code: -32603,
          message: 'Internal error',
          data: error.message
        },
        id: req.body?.id || null
      }
    };
  }
};
