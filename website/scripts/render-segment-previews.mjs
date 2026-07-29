#!/usr/bin/env node
// Renders a live prompt-preview SVG for every segment doc's own "Sample Configuration", so
// docs/segments/*.mdx can show "here is what this produces" above "here is how to write it"
// (see src/components/Config.js, which reads the output of this script).
//
// For each docs/segments/**/*.mdx file (skipping overview.mdx, which is an index page rather
// than a segment doc) this:
//   1. extracts the first <Config data={{ ... }}/> object literal from the doc source - the one
//      already verified to always sit under "## Sample Configuration" ("## Examples" for the one
//      exception, languages/language.mdx), with any later <Config/> blocks in the same doc
//      (illustrative examples elsewhere in the page) ignored;
//   2. wraps it as the lone segment of a minimal version-4 config, one "left"-aligned block;
//   3. shells out to the oh-my-posh CLI (see buildPoshArgs below - same flags/font metrics as
//      export_themes.mjs, so a segment preview and a theme-gallery render use an identical grid)
//      to export that config to SVG, fed by the same website/segment_data.json every theme
//      preview already renders from;
//   4. writes generated/segment-previews.json keyed by doc id (the frontmatter `id:`, exposed by
//      the doc page's own useDoc() as frontMatter.id - see Config.js), each entry a
//      { segment, svg } pair so Config.js can confirm which of a page's possibly-several
//      <Config/> blocks the preview actually belongs to (see the manifest-building comment in
//      main() below).
//
// Output is gitignored under /generated (see .gitignore), regenerated on every build exactly
// like generated/themes.json, and reached by Config.js via a plain static import - not
// usePluginData - for the same bundle-size reason themes.json documents on plugins/themes/index.js:
// dozens of inlined SVGs is enough weight that it must be scoped to the pages that actually render
// it (via webpack's own code-splitting) rather than landing in the shared main.js runtime chunk
// that every page on the site pays for, forever.
import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promises } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

import { asyncPool } from '../export_themes.mjs';
import { VICTOR_MONO } from '../font-metrics.mjs';

const execFileAsync = promisify(execFile);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const CONFIG = {
  DOCS_DIR: join(__dirname, '../docs/segments'),
  OUTPUT_FILE: join(__dirname, '../generated/segment-previews.json'),
  SEGMENT_DATA_FILE: join(__dirname, '../segment_data.json'),
  CONCURRENCY: 8,
  FONT_FAMILY: VICTOR_MONO.FONT_FAMILY,
  CELL_WIDTH: VICTOR_MONO.CELL_WIDTH,
  LINE_HEIGHT: VICTOR_MONO.LINE_HEIGHT,
  FILL_ASCENT: VICTOR_MONO.FILL_ASCENT,
  FILL_DESCENT: VICTOR_MONO.FILL_DESCENT,
};

// Same override hook export_themes.mjs offers: point at a locally built binary instead of
// whatever is on PATH.
const OMP_BIN = process.env.OMP_BIN || 'oh-my-posh';

// Doc files that are not themselves a segment (an index/overview page) and legitimately carry no
// <Config/> block to extract.
const PREVIEW_COLUMNS = 64;

const NOT_A_SEGMENT_DOC = new Set(['overview.mdx']);

function buildPoshArgs(configPath, outputPath) {
  return [
    'config',
    'export',
    'image',
    `--config=${configPath}`,
    `--output=${outputPath}`,
    `--font-family=${CONFIG.FONT_FAMILY}`,
    `--cell-width=${CONFIG.CELL_WIDTH}`,
    `--line-height=${CONFIG.LINE_HEIGHT}`,
    `--fill-ascent=${CONFIG.FILL_ASCENT}`,
    `--fill-descent=${CONFIG.FILL_DESCENT}`,
    // Narrower than the gallery's 120. A theme fills the width; one segment occupies a
    // fraction of it, so at 120 columns the preview is mostly empty terminal and the
    // segment renders small enough to squint at. 64 is the smallest width at which none of
    // the 117 samples wraps onto a second line - 48 wrapped five of them (argocd, gcp, nba,
    // orthodoxcal, pulumi), and a segment broken across two rows misrepresents it.
    `--terminal-width=${PREVIEW_COLUMNS}`,
    // Same rationale as export_themes.mjs: pin the render to the hand-written sample data so a
    // segment renders a plausible machine rather than whatever ran the build.
    `--data=${CONFIG.SEGMENT_DATA_FILE}`,
    '--data-only',
  ];
}

/**
 * Walks docs/segments/<group>/*.mdx, returning [{ group, fileName, filePath }, ...] sorted the
 * same way plugins/segments and extract-segment-properties.mjs already walk this tree.
 */
async function listSegmentDocs() {
  const groups = (await promises.readdir(CONFIG.DOCS_DIR, { withFileTypes: true }))
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort();

  const docs = [];
  for (const group of groups) {
    const groupDir = join(CONFIG.DOCS_DIR, group);
    const files = (await promises.readdir(groupDir, { withFileTypes: true }))
      .filter((entry) => entry.isFile() && entry.name.endsWith('.mdx'))
      .map((entry) => entry.name)
      .filter((name) => !NOT_A_SEGMENT_DOC.has(name))
      .sort();

    for (const fileName of files) {
      docs.push({ group, fileName, filePath: join(groupDir, fileName) });
    }
  }

  return docs;
}

/**
 * Reads the `id:` frontmatter field a doc's `---`-fenced header. This is the id Docusaurus'
 * useDoc() reports back for the page (see Config.js), which is why the output is keyed on it
 * rather than on the filename.
 */
function extractDocId(filePath, content) {
  const lines = content.split(/\r?\n/);

  if (lines[0] !== '---') {
    return { error: `does not start with a "---" frontmatter fence` };
  }

  let end = -1;
  for (let i = 1; i < lines.length; i += 1) {
    if (lines[i] === '---') {
      end = i;
      break;
    }
  }
  if (end === -1) {
    return { error: `has no closing "---" frontmatter fence` };
  }

  for (const line of lines.slice(1, end)) {
    const sep = line.indexOf(':');
    if (sep === -1) continue;
    const key = line.slice(0, sep).trim();
    if (key === 'id') {
      const id = line.slice(sep + 1).trim();
      if (!id) return { error: 'frontmatter "id" is empty' };
      return { id };
    }
  }

  return { error: 'frontmatter has no "id" field' };
}

/**
 * Scans forward from `openIndex` (which must point at a '{') for the matching close brace,
 * tracking string literals (single/double/template-quoted, backslash-escape aware) so that a
 * '{' or '}' inside a quoted template string - e.g. `"{{ .HEAD }}"` - is not mistaken for
 * structure. Returns the index of the matching '}', or -1 if the source ends unbalanced.
 */
function findMatchingBrace(source, openIndex) {
  let depth = 0;
  let inString = null;

  for (let i = openIndex; i < source.length; i += 1) {
    const ch = source[i];

    if (inString) {
      if (ch === '\\') {
        i += 1; // skip the escaped character, whatever it is
        continue;
      }
      if (ch === inString) inString = null;
      continue;
    }

    if (ch === '"' || ch === "'" || ch === '`') {
      inString = ch;
      continue;
    }

    if (ch === '{') {
      depth += 1;
    } else if (ch === '}') {
      depth -= 1;
      if (depth === 0) return i;
    }
  }

  return -1;
}

/**
 * Extracts the first <Config data={{ ... }}/> object literal from a doc's raw source and
 * evaluates it into a plain object. The literal is JS, not JSON (unquoted keys, ""
 * escapes, trailing commas, nested objects) - this is repo-controlled content already executed
 * as JSX at build time, so evaluating it here is no wider a trust boundary than MDX compilation
 * already crosses.
 */
function extractSegmentConfig(content) {
  const configIdx = content.indexOf('<Config');
  if (configIdx === -1) {
    return { error: 'no <Config/> element found' };
  }

  const dataIdx = content.indexOf('data=', configIdx);
  if (dataIdx === -1) {
    return { error: '<Config/> has no "data" prop' };
  }

  const braceStart = content.indexOf('{', dataIdx);
  if (braceStart === -1) {
    return { error: '"data" prop has no opening brace' };
  }

  const braceEnd = findMatchingBrace(content, braceStart);
  if (braceEnd === -1) {
    return { error: 'unbalanced braces in "data" prop' };
  }

  // content.slice(braceStart, braceEnd + 1) is the full '{{ ... }}': the JSX expression
  // container's own braces wrapped around the object literal. Stripping one layer off each end
  // leaves just the object literal text.
  const wrapped = content.slice(braceStart, braceEnd + 1);
  const literal = wrapped.slice(1, -1).trim();

  let segment;
  try {
    // eslint-disable-next-line no-new-func
    segment = new Function(`"use strict"; return (${literal});`)();
  } catch (error) {
    return { error: `failed to evaluate object literal: ${error.message}` };
  }

  if (segment === null || typeof segment !== 'object' || Array.isArray(segment)) {
    return { error: 'evaluated "data" is not a plain object' };
  }
  if (typeof segment.type !== 'string' || segment.type === '') {
    return { error: 'evaluated "data" has no string "type"' };
  }

  return { segment };
}

function buildPromptConfig(segment) {
  return {
    version: 4,
    final_space: true,
    blocks: [
      {
        type: 'prompt',
        alignment: 'left',
        segments: [segment],
      },
    ],
  };
}

async function renderSegmentPreview(doc) {
  const content = await promises.readFile(doc.filePath, 'utf8');

  const idResult = extractDocId(doc.filePath, content);
  if (idResult.error) {
    return { doc, error: `could not determine doc id: ${idResult.error}` };
  }

  const configResult = extractSegmentConfig(content);
  if (configResult.error) {
    return { doc, error: `could not extract sample config: ${configResult.error}` };
  }

  const promptConfig = buildPromptConfig(configResult.segment);

  const tmpConfigPath = join(tmpdir(), `omp-preview-config-${randomUUID()}.omp.json`);
  const tmpOutputPath = join(tmpdir(), `omp-preview-${randomUUID()}.svg`);

  try {
    await promises.writeFile(tmpConfigPath, JSON.stringify(promptConfig));

    let stderr;
    try {
      ({ stderr } = await execFileAsync(OMP_BIN, buildPoshArgs(tmpConfigPath, tmpOutputPath)));
    } catch (error) {
      return { doc, error: `CLI render failed: ${error.message}` };
    }

    if (stderr) {
      // Non-zero exit already returned above; incidental stderr (e.g. the hand-written-data-file
      // warning export_themes.mjs also swallows) must not fail the build.
      console.warn(`${doc.fileName}: ${stderr.trim()}`);
    }

    const svg = await promises.readFile(tmpOutputPath, 'utf8');

    console.info(`Rendered preview for ${idResult.id} (${doc.group}/${doc.fileName})`);

    return { doc, id: idResult.id, segment: configResult.segment, svg };
  } finally {
    await promises.unlink(tmpConfigPath).catch(() => {});
    await promises.unlink(tmpOutputPath).catch(() => {});
  }
}

async function ensureDirectories() {
  await promises.mkdir(dirname(CONFIG.OUTPUT_FILE), { recursive: true });
}

async function main() {
  try {
    console.log('Starting segment preview render process...');

    await ensureDirectories();

    const docs = await listSegmentDocs();
    console.log(`Found ${docs.length} segment docs to process`);

    const results = [];
    for await (const result of asyncPool(CONFIG.CONCURRENCY, docs, renderSegmentPreview)) {
      results.push(result);
    }

    const rendered = results.filter((result) => !result.error);
    const failed = results.filter((result) => result.error);

    // { segment, svg } rather than a bare svg string: six segment docs (see module comment)
    // render a second, illustrative <Config/> elsewhere on the page with the same doc id but a
    // different sample object. Config.js only shows the preview above the <Config/> whose own
    // `data` prop deep-equals `segment` here, so the generic preview never appears next to an
    // unrelated example further down the same page.
    const manifest = {};
    for (const result of rendered) {
      manifest[result.id] = { segment: result.segment, svg: result.svg };
    }

    await promises.writeFile(CONFIG.OUTPUT_FILE, JSON.stringify(manifest));

    console.log(`Successfully rendered ${rendered.length}/${docs.length} segment previews to ${CONFIG.OUTPUT_FILE}`);

    if (failed.length > 0) {
      console.warn(`${failed.length} doc(s) could not be rendered:`);
      for (const result of failed) {
        console.warn(`  - ${result.doc.group}/${result.doc.fileName}: ${result.error}`);
      }
    }
  } catch (error) {
    console.error('Segment preview render process failed:', error.message);
    process.exit(1);
  }
}

if (process.argv[1] === __filename) {
  main();
}

export {
  listSegmentDocs,
  extractDocId,
  extractSegmentConfig,
  findMatchingBrace,
  buildPromptConfig,
  renderSegmentPreview,
  main,
};
