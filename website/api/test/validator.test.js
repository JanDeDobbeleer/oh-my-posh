const { describe, it } = require('node:test');
const assert = require('node:assert');
const validator = require('../shared/validator.js');

// Test fixtures
const fixtures = {
  validConfig: `{
    "version": 3,
    "blocks": [
      {
        "type": "prompt",
        "alignment": "left",
        "segments": [
          {
            "type": "path",
            "style": "powerline",
            "foreground": "#ffffff",
            "background": "#0077c2",
            "template": " {{ .Path }} "
          }
        ]
      }
    ]
  }`,

  validYamlConfig: `
version: 3
blocks:
  - type: prompt
    alignment: left
    segments:
      - type: path
        style: powerline
        foreground: "#ffffff"
        background: "#0077c2"
        template: " {{ .Path }} "
  `,

  invalidConfig: `{
    "version": 3,
    "blocks": []
  }`,

  validSegment: `{
    "type": "git",
    "style": "powerline",
    "foreground": "#ffffff",
    "background": "#007acc",
    "template": " {{ .HEAD }} "
  }`,

  invalidSegmentNoType: `{
    "style": "powerline"
  }`,

  invalidSegmentNoStyle: `{
    "type": "git"
  }`,

  malformedJson: `{ "invalid": json }`
};

describe('oh-my-posh validator', () => {
  describe('validateConfig', () => {
    it('should validate a valid JSON config', async () => {
      const result = await validator.validateConfig(fixtures.validConfig, 'json');
      
      assert.strictEqual(result.valid, true, 'Config should be valid');
      assert.strictEqual(result.errors.length, 0, 'Should have no errors');
      assert.strictEqual(result.detectedFormat, 'json');
      assert.ok(result.parsedConfig, 'Should have parsed config');
    });

    it('should validate a valid YAML config', async () => {
      const result = await validator.validateConfig(fixtures.validYamlConfig, 'yaml');
      
      assert.strictEqual(result.valid, true, 'Config should be valid');
      assert.strictEqual(result.errors.length, 0, 'Should have no errors');
      assert.strictEqual(result.detectedFormat, 'yaml');
    });

    it('should auto-detect format', async () => {
      const result = await validator.validateConfig(fixtures.validConfig, 'auto');
      
      assert.strictEqual(result.valid, true);
      assert.strictEqual(result.detectedFormat, 'json');
    });

    it('should handle malformed JSON', async () => {
      const result = await validator.validateConfig(fixtures.malformedJson, 'json');
      
      assert.strictEqual(result.valid, false, 'Config should be invalid');
      assert.ok(result.errors.length > 0, 'Should have parse errors');
    });

    it('should include parsed config in result', async () => {
      const result = await validator.validateConfig(fixtures.validConfig, 'json');
      
      assert.ok(result.parsedConfig);
      assert.strictEqual(result.parsedConfig.version, 3);
      assert.ok(Array.isArray(result.parsedConfig.blocks));
    });
  });

  describe('validateSegment', () => {
    it('should validate a valid segment', async () => {
      const result = await validator.validateSegment(fixtures.validSegment, 'json');
      
      assert.strictEqual(result.valid, true, 'Segment should be valid');
      assert.strictEqual(result.errors.length, 0, 'Should have no errors');
      assert.ok(result.parsedSegment, 'Should have parsed segment');
    });

    it('should reject segment missing type', async () => {
      const result = await validator.validateSegment(fixtures.invalidSegmentNoType, 'json');
      
      assert.strictEqual(result.valid, false, 'Segment should be invalid');
      assert.ok(result.errors.length > 0, 'Should have errors');
      assert.ok(
        result.errors.some(e => e.path === 'type'),
        'Should have error about missing type'
      );
    });

    it('should reject segment missing style', async () => {
      const result = await validator.validateSegment(fixtures.invalidSegmentNoStyle, 'json');
      
      assert.strictEqual(result.valid, false, 'Segment should be invalid');
      assert.ok(result.errors.length > 0, 'Should have errors');
      assert.ok(
        result.errors.some(e => e.path === 'style'),
        'Should have error about missing style'
      );
    });

    it('should include parsed segment in result', async () => {
      const result = await validator.validateSegment(fixtures.validSegment, 'json');
      
      assert.ok(result.parsedSegment);
      assert.strictEqual(result.parsedSegment.type, 'git');
      assert.strictEqual(result.parsedSegment.style, 'powerline');
    });
  });

  describe('parseConfig', () => {
    it('should parse JSON', () => {
      const result = validator.parseConfig(fixtures.validConfig, 'json');
      
      assert.ok(result);
      assert.strictEqual(result.version, 3);
    });

    it('should parse YAML', () => {
      const result = validator.parseConfig(fixtures.validYamlConfig, 'yaml');
      
      assert.ok(result);
      assert.strictEqual(result.version, 3);
    });

    it('should throw on invalid content', () => {
      assert.throws(
        () => validator.parseConfig(fixtures.malformedJson, 'json'),
        /Failed to parse/
      );
    });
  });

  describe('detectFormat', () => {
    it('should detect JSON format', () => {
      const format = validator.detectFormat('{ "key": "value" }');
      assert.strictEqual(format, 'json');
    });

    it('should detect YAML format', () => {
      const format = validator.detectFormat('key: value\nother: thing');
      assert.strictEqual(format, 'yaml');
    });
  });

  describe('formatErrors', () => {
    it('should format validation errors', () => {
      const errors = [
        {
          instancePath: '/blocks/0',
          message: 'must have required property \'segments\'',
          keyword: 'required',
          params: { missingProperty: 'segments' }
        }
      ];

      const formatted = validator.formatErrors(errors);
      
      assert.ok(Array.isArray(formatted));
      assert.strictEqual(formatted.length, 1);
      assert.strictEqual(formatted[0].path, '/blocks/0');
      assert.ok(formatted[0].message.includes('segments'));
    });

    it('should handle empty errors array', () => {
      const formatted = validator.formatErrors([]);
      assert.deepStrictEqual(formatted, []);
    });
  });
});
