const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const yaml = require('js-yaml');
const toml = require('@iarna/toml');
const fs = require('fs');
const path = require('path');

// Load the schema - will be copied during build
let schema;
try {
  // Try to load from the static directory (after build)
  const schemaPath = path.join(__dirname, '..', '..', 'static', 'schema.json');
  if (fs.existsSync(schemaPath)) {
    schema = JSON.parse(fs.readFileSync(schemaPath, 'utf8'));
  } else {
    // Fallback to relative path during development
    const devSchemaPath = path.join(__dirname, '..', '..', '..', 'themes', 'schema.json');
    schema = JSON.parse(fs.readFileSync(devSchemaPath, 'utf8'));
  }
} catch (error) {
  console.error('Failed to load schema:', error);
  schema = null;
}

// Initialize AJV with schema
const ajv = new Ajv({
  allErrors: true,
  verbose: true,
  strict: false,
  validateFormats: true
});
addFormats(ajv);

// Add custom format for color validation
ajv.addFormat('color', {
  validate: (data) => {
    if (typeof data !== 'string') return false;
    // This is a simple validation - the schema pattern handles the real validation
    return true;
  }
});

let validate;
if (schema) {
  try {
    validate = ajv.compile(schema);
  } catch (error) {
    console.error('Failed to compile schema:', error);
  }
}

/**
 * Detect the format of the configuration content
 * @param {string} content - The configuration content
 * @returns {string} The detected format (json, yaml, or toml)
 */
function detectFormat(content) {
  const trimmed = content.trim();
  
  // Try JSON first
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    return 'json';
  }
  
  // Check for TOML indicators
  if (trimmed.match(/^\[.*\]$/m) || trimmed.match(/^[a-zA-Z_][a-zA-Z0-9_]*\s*=/m)) {
    return 'toml';
  }
  
  // Default to YAML (most permissive)
  return 'yaml';
}

/**
 * Parse configuration content based on format
 * @param {string} content - The configuration content
 * @param {string} format - The format (json, yaml, toml, or auto)
 * @returns {Object} Parsed configuration object
 */
function parseConfig(content, format) {
  if (!content || typeof content !== 'string') {
    throw new Error('Content must be a non-empty string');
  }

  const detectedFormat = format === 'auto' ? detectFormat(content) : format;

  try {
    switch (detectedFormat.toLowerCase()) {
      case 'json':
        return JSON.parse(content);
      
      case 'yaml':
      case 'yml':
        return yaml.load(content);
      
      case 'toml':
        return toml.parse(content);
      
      default:
        throw new Error(`Unsupported format: ${detectedFormat}`);
    }
  } catch (error) {
    throw new Error(`Failed to parse ${detectedFormat}: ${error.message}`);
  }
}

/**
 * Format validation errors into human-readable messages
 * @param {Array} errors - AJV validation errors
 * @returns {Array} Formatted error messages
 */
function formatErrors(errors) {
  if (!errors || errors.length === 0) {
    return [];
  }

  return errors.map(error => {
    const path = error.instancePath || 'root';
    let message = error.message;

    // Enhance error messages based on error type
    switch (error.keyword) {
      case 'required':
        message = `Missing required property: ${error.params.missingProperty}`;
        break;
      case 'enum':
        message = `Value must be one of: ${error.params.allowedValues.join(', ')}`;
        break;
      case 'type':
        message = `Must be of type ${error.params.type}`;
        break;
      case 'pattern':
        message = `Must match pattern: ${error.params.pattern}`;
        break;
      case 'additionalProperties':
        message = `Unexpected property: ${error.params.additionalProperty}`;
        break;
    }

    return {
      path: path,
      message: message,
      keyword: error.keyword,
      params: error.params,
      data: error.data
    };
  });
}

/**
 * Validate an oh-my-posh configuration
 * @param {string} content - The configuration content
 * @param {string} format - The format (json, yaml, toml, or auto)
 * @returns {Object} Validation result
 */
async function validateConfig(content, format = 'auto') {
  const result = {
    valid: false,
    errors: [],
    warnings: [],
    detectedFormat: null,
    parsedConfig: null
  };

  // Check if schema is loaded
  if (!schema || !validate) {
    result.errors.push({
      path: 'schema',
      message: 'Schema could not be loaded. Validation is not available.',
      keyword: 'schema',
      params: {},
      data: null
    });
    return result;
  }

  try {
    // Parse the configuration
    const detectedFormat = format === 'auto' ? detectFormat(content) : format;
    result.detectedFormat = detectedFormat;
    
    const config = parseConfig(content, format);
    result.parsedConfig = config;

    // Validate against schema
    const isValid = validate(config);
    result.valid = isValid;

    if (!isValid && validate.errors) {
      result.errors = formatErrors(validate.errors);
    }

    // Add warnings for common issues
    if (config && typeof config === 'object') {
      // Check for deprecated version
      if (config.version && config.version < 2) {
        result.warnings.push({
          path: 'version',
          message: 'Using deprecated version format. Consider upgrading to version 2 or 3.',
          type: 'deprecation'
        });
      }

      // Check for missing $schema
      if (!config.$schema) {
        result.warnings.push({
          path: '$schema',
          message: 'Consider adding "$schema" property for better editor support.',
          type: 'recommendation'
        });
      }
    }

  } catch (error) {
    result.valid = false;
    result.errors.push({
      path: 'parse',
      message: error.message,
      keyword: 'parse',
      params: {},
      data: null
    });
  }

  return result;
}

module.exports = {
  validateConfig,
  parseConfig,
  detectFormat,
  formatErrors
};
