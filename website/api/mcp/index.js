const validator = require('../shared/validator.js');

const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Accept',
};

const TOOLS = [
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
  },
  {
    name: 'validate_segment',
    description: 'Validate a segment snippet against the oh-my-posh schema. Useful for validating individual prompt segments before adding them to a configuration.',
    inputSchema: {
      type: 'object',
      properties: {
        content: {
          type: 'string',
          description: 'The segment content as a string (JSON, YAML, or TOML)'
        },
        format: {
          type: 'string',
          enum: ['json', 'yaml', 'toml', 'auto'],
          description: 'The format of the segment (auto-detect if not specified)',
          default: 'auto'
        }
      },
      required: ['content']
    }
  }
];

/**
 * Azure Function entry point for MCP server
 */
module.exports = async function (context, req) {
  context.log('MCP validator function processed a request');

  // Handle OPTIONS preflight requests
  if (req.method === 'OPTIONS') {
    context.res = {
      status: 204,
      headers: CORS_HEADERS,
      body: '',
    };
    return;
  }

  // Handle GET requests - return server info
  if (req.method === 'GET') {
    context.res = {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        ...CORS_HEADERS,
      },
      body: {
        name: 'oh-my-posh-validator',
        version: '1.0.0',
        description: 'MCP server for validating oh-my-posh configurations',
        capabilities: {
          tools: {}
        },
        tools: TOOLS
      }
    };
    return;
  }

  // Handle POST requests - process MCP protocol messages
  let message = req.body;
  try {
    // Parse the body if it's a string
    if (typeof message === 'string') {
      try {
        message = JSON.parse(message);
      } catch (e) {
        context.log.error('Failed to parse request body as JSON:', e);
        context.res = {
          status: 400,
          headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
          body: {
            jsonrpc: '2.0',
            error: {code: -32700, message: 'Parse error: Invalid JSON'},
            id: null,
          },
        };
        return;
      }
    }

    if (!message || !message.jsonrpc || message.jsonrpc !== '2.0') {
      context.log('Invalid JSON-RPC message:', message);
      context.res = {
        status: 400,
        headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
        body: {
          jsonrpc: '2.0',
          error: {
            code: -32600,
            message: 'Invalid Request: Not a valid JSON-RPC 2.0 message',
          },
          id: message?.id || null,
        },
      };
      return;
    }

    context.log('Received message:', JSON.stringify({
      jsonrpc: message.jsonrpc,
      method: message.method,
      id: message.id,
      ...(message.params?.name && { toolName: message.params.name })
    }));

    // Handle list tools request
    if (message.method === 'tools/list') {
      context.res = {
        status: 200,
        headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
        body: {
          jsonrpc: '2.0',
          result: {
            tools: TOOLS
          },
          id: message.id,
        },
      };
      return;
    }

    // Handle tool call request
    if (message.method === 'tools/call') {
      if (!message.params?.name || !message.params?.arguments?.content) {
        context.res = {
          status: 400,
          headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
          body: {
            jsonrpc: '2.0',
            error: {code: -32602, message: 'Invalid params: name and arguments.content are required'},
            id: message.id,
          },
        };
        return;
      }

      const { name, arguments: args } = message.params;

      let result;
      
      if (name === 'validate_config') {
        result = await validator.validateConfig(args.content, args.format || 'auto');
      } else if (name === 'validate_segment') {
        result = await validator.validateSegment(args.content, args.format || 'auto');
      } else {
        context.res = {
          status: 200,
          headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
          body: {
            jsonrpc: '2.0',
            error: {code: -32601, message: `Unknown tool: ${name}`},
            id: message.id,
          },
        };
        return;
      }

      context.res = {
        status: 200,
        headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
        body: {
          jsonrpc: '2.0',
          result: {
            content: [{type: 'text', text: JSON.stringify(result, null, 2)}],
            isError: !result.valid,
          },
          id: message.id,
        },
      };
      return;
    }

    // Handle initialize request
    if (message.method === 'initialize') {
      context.res = {
        status: 200,
        headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
        body: {
          jsonrpc: '2.0',
          result: {
            protocolVersion: '2024-11-05',
            capabilities: {tools: {}},
            serverInfo: {name: 'oh-my-posh-validator', version: '1.0.0'},
          },
          id: message.id,
        },
      };
      return;
    }

    // Unknown method
    context.res = {
      status: 200,
      headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
      body: {
        jsonrpc: '2.0',
        error: {code: -32601, message: `Method not found: ${message.method}`},
        id: message.id,
      },
    };

  } catch (error) {
    context.log.error('Error processing MCP request:', error);
    context.res = {
      status: 500,
      headers: {'Content-Type': 'application/json', ...CORS_HEADERS},
      body: {
        jsonrpc: '2.0',
        error: {
          code: -32603,
          message: 'Internal error',
        },
        id: message?.id || null,
      },
    };
  }
};
