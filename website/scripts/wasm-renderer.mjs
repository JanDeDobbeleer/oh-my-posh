// Renders a config to SVG by running the studio's own WebAssembly module under Node, instead of
// spawning the oh-my-posh CLI once per render.
//
// Same code either way: the module and the CLI both call render.SVG, so the bytes are identical.
// What changes is the cost. Building the gallery used to spawn 124 processes and the segment
// previews another 117; a browser instantiates this module once and renders every prompt in it,
// and so can a build script.
//
// It is also what lets the SVG encoder leave the CLI entirely: nothing outside this module needs
// to be able to draw one.
import { readFile } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const WASM = join(WEBSITE_DIR, 'static', 'omp.wasm');
const WASM_EXEC = join(WEBSITE_DIR, 'static', 'wasm_exec.js');

let ready;

// Go's own host shim declares `globalThis.Go` and expects a browser-ish global scope. It is
// written as a classic script rather than a module, so it is evaluated rather than imported.
async function loadRuntime() {
  if (globalThis.Go) {
    return;
  }

  const shim = await readFile(WASM_EXEC, 'utf8');
  // eslint-disable-next-line no-new-func
  new Function(shim).call(globalThis);

  if (!globalThis.Go) {
    throw new Error(`${WASM_EXEC} did not define globalThis.Go. Run \`npm run wasm\` to rebuild it.`);
  }
}

// One instance for the whole process. The module's main() parks on a channel once it has
// published render(), so the instance stays alive and every later call reuses it - which is the
// entire point of doing this in-process.
async function instantiate() {
  await loadRuntime();

  const go = new globalThis.Go();
  const { instance } = await WebAssembly.instantiate(await readFile(WASM), go.importObject);

  // Deliberately not awaited: go.run resolves only when main() returns, which it never does.
  // The exports are published synchronously before it parks, so render is callable on the next
  // line - the same contract the browser loader relies on.
  go.run(instance).catch((err) => {
    throw err;
  });

  if (typeof globalThis.render !== 'function') {
    throw new Error('the wasm module did not publish render(). Run `npm run wasm` to rebuild it.');
  }

  return globalThis.render;
}

/**
 * Renders one config to an SVG document.
 *
 * @param {string} config - the config text.
 * @param {string} format - json, yaml or toml; the same strings config.ParseBytes takes.
 * @param {string} data - a template data file's contents, as JSON.
 * @param {object} options - columns plus the font metrics (see render.SVGOptions).
 * @returns {Promise<string>} the SVG document.
 */
export async function renderSVG(config, format, data, options) {
  ready ??= instantiate();

  const render = await ready;
  const result = render(config, format, data, options);

  // The module reports failure as { error }, never by throwing: a Go panic would otherwise take
  // the whole instance down and every later render with it.
  if (result.error) {
    throw new Error(result.error);
  }

  return result.svg;
}
