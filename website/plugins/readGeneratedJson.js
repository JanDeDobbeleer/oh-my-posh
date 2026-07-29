/**
 * Shared load-and-validate scaffold for the build-generated JSON files plugins/segments and
 * plugins/themes each guard: does the file exist, can it be read, does it parse. Each plugin still
 * owns its own per-entry shape/duplicate checks afterwards - those differ too much between a
 * registry keyed on docId and a manifest keyed on theme name to be worth folding in here too.
 *
 * Most of these files are arrays, so readGeneratedArray adds that check; generated/hero.json is a
 * single theme entry (see export_themes.mjs) and uses the plain reader.
 */

const fs = require('fs');

/**
 * @param {string} filePath - absolute path to the generated JSON file.
 * @param {string} pluginName - the docusaurus plugin name, used to prefix every error.
 * @param {string} notFoundHint - one or two sentences telling the developer what generates this
 *   file and how to (re)generate it, appended to the "missing <filePath>." error.
 * @returns {unknown} the parsed contents.
 */
function readGeneratedJson(filePath, pluginName, notFoundHint) {
  if (!fs.existsSync(filePath)) {
    throw new Error(`${pluginName} plugin: missing ${filePath}. ${notFoundHint}`);
  }

  let raw;
  try {
    raw = fs.readFileSync(filePath, 'utf8');
  } catch (err) {
    throw new Error(`${pluginName} plugin: failed to read ${filePath}: ${err.message}`);
  }

  let data;
  try {
    data = JSON.parse(raw);
  } catch (err) {
    throw new Error(`${pluginName} plugin: ${filePath} is not valid JSON: ${err.message}`);
  }

  return data;
}

/**
 * readGeneratedJson, plus the top-level-is-an-array check every generated file but hero.json needs.
 *
 * @returns {unknown[]} the parsed top-level array.
 */
function readGeneratedArray(filePath, pluginName, notFoundHint) {
  const data = readGeneratedJson(filePath, pluginName, notFoundHint);

  if (!Array.isArray(data)) {
    throw new Error(`${pluginName} plugin: ${filePath} must contain a JSON array.`);
  }

  return data;
}

module.exports = { readGeneratedJson, readGeneratedArray };
