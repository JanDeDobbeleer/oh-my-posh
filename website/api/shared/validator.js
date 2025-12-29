const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const yaml = require('js-yaml');
const toml = require('@iarna/toml');
const fs = require('fs');
const path = require('path');
const axios = require('axios');

// Configuration constants
const SCHEMA_GITHUB_URL = 'https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json';
const SCHEMA_FETCH_TIMEOUT = 10000;

// Schema cache
let schema = null;
let schemaLoadPromise = null;

/**
 * Load the schema from the local data folder, with GitHub fallback
 * @returns {Promise<Object>} The loaded schema
 */
async function loadSchema() {
  // Return cached schema if available
  if (schema) {
    return schema;
  }

  // Return existing promise if schema is already being loaded
  if (schemaLoadPromise) {
    return schemaLoadPromise;
  }

  schemaLoadPromise = (async () => {
    // Try loading from local data directory first
    try {
      const schemaPath = path.join(__dirname, '..', 'data', 'schema.json');
      console.log('Attempting to load schema from:', schemaPath);

      if (fs.existsSync(schemaPath)) {
        const loadedSchema = JSON.parse(fs.readFileSync(schemaPath, 'utf8'));
        console.log('Schema loaded successfully from local data folder');
        schema = loadedSchema; // Set cache after successful load
        return loadedSchema;
      } else {
        console.log('Local schema file not found, will fetch from GitHub');
      }
    } catch (error) {
      console.log('Failed to load schema from local data folder:', error.message);
    }

    // Fallback to GitHub
    try {
      console.log('Fetching schema from GitHub');
      const response = await axios.get(SCHEMA_GITHUB_URL, {
        timeout: SCHEMA_FETCH_TIMEOUT
      });
      const loadedSchema = response.data;
      console.log('Schema loaded successfully from GitHub');
      schema = loadedSchema; // Set cache after successful load
      return loadedSchema;
    } catch (error) {
      console.error('Failed to fetch schema from GitHub:', error.message);
      throw new Error('Could not load schema from local data folder or GitHub');
    }
  })()
    .finally(() => {
      // Reset promise to allow retry on subsequent calls if this attempt failed
      schemaLoadPromise = null;
    });

  return schemaLoadPromise;
}

// Initialize AJV
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

// Compile validator from schema
let validate = null;
async function getValidator() {
  if (validate) {
    return validate;
  }

  try {
    const loadedSchema = await loadSchema();
    if (!loadedSchema) {
      throw new Error('Schema loading returned null');
    }

    validate = ajv.compile(loadedSchema);
    return validate;
  } catch (error) {
    console.error('Failed to load or compile schema:', error);
    throw error; // Propagate error instead of returning null
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
 * @returns {Promise<Object>} Validation result
 */
async function validateConfig(content, format = 'auto') {
  const result = {
    valid: false,
    errors: [],
    warnings: [],
    detectedFormat: null,
    parsedConfig: null
  };

  try {
    // Load and compile validator
    const validator = await getValidator();

    // Parse the configuration
    const detectedFormat = format === 'auto' ? detectFormat(content) : format;
    result.detectedFormat = detectedFormat;

    const config = parseConfig(content, format);
    result.parsedConfig = config;

    // Validate against schema
    const isValid = validator(config);
    result.valid = isValid;

    if (!isValid && validator.errors) {
      result.errors = formatErrors(validator.errors);
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

    // Check if it's a schema loading error
    if (error.message && error.message.includes('Could not load schema')) {
      result.errors.push({
        path: 'schema',
        message: 'Schema could not be loaded. Validation is not available.',
        keyword: 'schema',
        params: {},
        data: null
      });
    } else {
      result.errors.push({
        path: 'parse',
        message: error.message,
        keyword: 'parse',
        params: {},
        data: null
      });
    }
  }

  return result;
}

/**
 * Validate a segment configuration
 * @param {string} content - The segment content
 * @param {string} format - The format (json, yaml, toml, or auto)
 * @returns {Promise<Object>} Validation result
 */
async function validateSegment(content, format = 'auto') {
  const result = {
    valid: false,
    errors: [],
    warnings: [],
    detectedFormat: null,
    parsedSegment: null
  };

  try {
    // Load and compile validator
    const validator = await getValidator();

    // Parse the segment
    const detectedFormat = format === 'auto' ? detectFormat(content) : format;
    result.detectedFormat = detectedFormat;

    const segment = parseConfig(content, format);
    result.parsedSegment = segment;

    // Validate that it's an object
    if (!segment || typeof segment !== 'object' || Array.isArray(segment)) {
      result.errors.push({
        path: 'root',
        message: 'Segment must be a JSON object',
        keyword: 'type',
        params: {},
        data: segment
      });
      return result;
    }

    // Check required fields
    if (!segment.type) {
      result.errors.push({
        path: 'type',
        message: 'Missing required property: type',
        keyword: 'required',
        params: { missingProperty: 'type' },
        data: segment
      });
    }

    if (!segment.style) {
      result.errors.push({
        path: 'style',
        message: 'Missing required property: style',
        keyword: 'required',
        params: { missingProperty: 'style' },
        data: segment
      });
    }

    // If we already have errors, don't continue with schema validation
    if (result.errors.length > 0) {
      return result;
    }

    // Wrap segment in a minimal valid config for schema validation
    const wrappedConfig = {
      "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
      "version": 3,
      "blocks": [
        {
          "type": "prompt",
          "alignment": "left",
          "segments": [segment]
        }
      ]
    };

    // Validate against schema
    const isValid = validator(wrappedConfig);

    if (!isValid && validator.errors) {
      // Filter errors to only those related to the segment
      // Errors will have paths like /blocks/0/segments/0/...
      // Also filter out generic "if/then" schema errors as they're not helpful
      const segmentErrors = validator.errors.filter(error => {
        const isSegmentError = error.instancePath && error.instancePath.startsWith('/blocks/0/segments/0');
        const isIfThenError = error.keyword === 'if';
        return isSegmentError && !isIfThenError;
      });

      if (segmentErrors.length > 0) {
        result.errors = formatErrors(segmentErrors.map(error => ({
          ...error,
          // Clean up the path to make it relative to the segment
          instancePath: error.instancePath.replace('/blocks/0/segments/0', '')
        })));
      } else {
        // All segments passed validation
        result.valid = true;
      }
    } else {
      result.valid = true;
    }

    // Add specific warnings for segment
    if (segment.properties && segment.options) {
      result.warnings.push({
        path: 'properties',
        message: 'Both "properties" and "options" are present. "properties" is deprecated, use "options" instead.',
        type: 'deprecation'
      });
    } else if (segment.properties) {
      result.warnings.push({
        path: 'properties',
        message: 'The "properties" field is deprecated. Please rename it to "options".',
        type: 'deprecation'
      });
    }

  } catch (error) {
    result.valid = false;

    // Check if it's a schema loading error
    if (error.message && error.message.includes('Could not load schema')) {
      result.errors.push({
        path: 'schema',
        message: 'Schema could not be loaded. Validation is not available.',
        keyword: 'schema',
        params: {},
        data: null
      });
    } else {
      result.errors.push({
        path: 'parse',
        message: error.message,
        keyword: 'parse',
        params: {},
        data: null
      });
    }
  }

  return result;
}

module.exports = {
  validateConfig,
  validateSegment,
  parseConfig,
  detectFormat,
  formatErrors
};
