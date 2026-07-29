#!/usr/bin/env node
// Parses the "### Properties" table out of every segment doc (docs/segments/*/*.mdx) into
// { [segment type]: { [templateField]: goType } }. This is the schema that:
//   - website/segment_data.json's `segments` map is filled against (the keys are template
//     fields like `.Context`, never config options like `context_aliases`);
//   - the Go guard test (src/config, see segment_data_test.go or similar) reads to check that
//     every key in every segment_data.json entry is actually a documented property.
//
// Reused by both: run directly for a human-readable report, or `import { extractAll }` for the
// parsed data.
//
// docId vs type: registry.json's `docId` is the doc's filename stem (and part of its URL);
// `type` is the config `type:` value used as the segment_data.json key. They differ for two
// segments (go/golang, copilot_cli/copilot-cli) - this script always keys its output by `type`.
import { readFileSync, readdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const DOCS_ROOT = join(WEBSITE_DIR, 'docs', 'segments');
const REGISTRY_PATH = join(WEBSITE_DIR, 'plugins', 'segments', 'registry.json');

// Segment docs with no "### Properties" table at all - legitimately nothing to extract because
// the segment has no template fields (system/root.mdx: the segment's output is a fixed icon,
// nothing else). Verified by hand; keep in sync if a doc gains fields later - the report below
// will start listing it as unparsed the moment it does, since it's not silently ignored, only
// this one specific known-empty id is.
const KNOWN_NO_PROPERTIES = new Set([
  'root', // fixed icon output, no template fields at all
  'text', // doc says outright: "Text segments have no special options"
]);

function buildDocIndex() {
  const index = new Map(); // docId -> absolute file path
  const groups = readdirSync(DOCS_ROOT, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name);

  for (const group of groups) {
    const groupDir = join(DOCS_ROOT, group);
    const files = readdirSync(groupDir, { withFileTypes: true })
      .filter((e) => e.isFile() && e.name.endsWith('.mdx'))
      .map((e) => e.name)
      .filter((name) => name !== 'overview.mdx');

    for (const fileName of files) {
      const docId = fileName.slice(0, -'.mdx'.length);
      index.set(docId, join(groupDir, fileName));
    }
  }

  return index;
}

// Strips markdown inline formatting from a table cell: `code` spans, [text](url) and
// [text][ref] links, and surrounding whitespace.
function cleanCell(cell) {
  let text = cell.trim();
  text = text.replace(/`([^`]*)`/g, '$1');
  text = text.replace(/\[([^\]]+)\]\([^)]*\)/g, '$1');
  text = text.replace(/\[([^\]]+)\]\[[^\]]*\]/g, '$1');
  return text.trim();
}

function splitRow(line) {
  let trimmed = line.trim();
  if (trimmed.startsWith('|')) trimmed = trimmed.slice(1);
  if (trimmed.endsWith('|')) trimmed = trimmed.slice(0, -1);
  return trimmed.split('|').map((c) => c.trim());
}

function isSeparatorRow(cells) {
  return cells.length > 0 && cells.every((c) => /^:?-+:?$/.test(c.trim()));
}

// Extracts the top-level "### Properties" table (the one describing the segment's own template
// fields). Deliberately does NOT descend into deeper "#### Foo" sub-tables that document nested
// struct types (e.g. git's "#### Status", "#### Commit") - those describe fields of a field, not
// segment-level template properties, and segment_data.json / the guard test only need the
// top-level names.
function extractProperties(filePath) {
  const content = readFileSync(filePath, 'utf8');
  const lines = content.split(/\r?\n/);

  const headingIndex = lines.findIndex((l) => l.trim() === '### Properties');
  if (headingIndex === -1) {
    return { error: 'no "### Properties" heading found' };
  }

  // Find the table: skip blank lines, then expect a header row immediately followed by a
  // separator row.
  let i = headingIndex + 1;
  while (i < lines.length && lines[i].trim() === '') i += 1;

  if (i >= lines.length || !lines[i].trim().startsWith('|')) {
    return { error: 'no table found directly under "### Properties"' };
  }

  const headerCells = splitRow(lines[i]).map((c) => c.toLowerCase());
  const nameCol = headerCells.indexOf('name');
  const typeCol = headerCells.indexOf('type');
  if (nameCol === -1 || typeCol === -1) {
    return { error: `unrecognised table header: ${JSON.stringify(splitRow(lines[i]))}` };
  }

  i += 1;
  if (i >= lines.length || !isSeparatorRow(splitRow(lines[i]))) {
    return { error: 'table header not followed by a "---" separator row' };
  }
  i += 1;

  const fields = {};
  const order = [];
  let sawRow = false;
  for (; i < lines.length; i += 1) {
    const line = lines[i];
    if (!line.trim().startsWith('|')) break;
    const cells = splitRow(line);
    if (cells.length <= Math.max(nameCol, typeCol)) {
      return { error: `row has too few columns: ${JSON.stringify(cells)}` };
    }

    let name = cleanCell(cells[nameCol]);
    const type = cleanCell(cells[typeCol]);

    // Every doc but one (cloud/sitecore.mdx) writes the Name column as `.Field`, matching how
    // it's referenced in a template. sitecore.mdx omits the leading dot ("EndpointName", not
    // ".EndpointName") - strip it when present, don't require it.
    if (name.startsWith('.')) {
      name = name.slice(1);
    }
    if (name === '' || /\s/.test(name)) {
      return { error: `unparseable property name: ${JSON.stringify(cells[nameCol])}` };
    }

    if (fields[name] !== undefined && fields[name] !== type) {
      return { error: `duplicate property ".${name}" with conflicting types "${fields[name]}" vs "${type}"` };
    }
    fields[name] = type;
    if (!order.includes(name)) order.push(name);
    sawRow = true;
  }

  if (!sawRow) {
    return { error: 'table had a header but no data rows' };
  }

  return { fields, order };
}

export function extractAll() {
  const registry = JSON.parse(readFileSync(REGISTRY_PATH, 'utf8'));
  const docIndex = buildDocIndex();

  const result = {}; // type -> { field: type }
  const fieldOrder = {}; // type -> [field, ...] in doc order
  const problems = []; // { type, docId, reason }
  const knownEmptyUnused = new Set(KNOWN_NO_PROPERTIES);

  for (const entry of registry) {
    const filePath = docIndex.get(entry.docId);
    if (!filePath || !existsSync(filePath)) {
      problems.push({ type: entry.type, docId: entry.docId, reason: `doc file not found for docId "${entry.docId}"` });
      continue;
    }

    const parsed = extractProperties(filePath);
    if (parsed.error) {
      if (KNOWN_NO_PROPERTIES.has(entry.docId)) {
        knownEmptyUnused.delete(entry.docId);
        result[entry.type] = {};
        fieldOrder[entry.type] = [];
        continue;
      }
      problems.push({ type: entry.type, docId: entry.docId, reason: parsed.error });
      continue;
    }

    result[entry.type] = parsed.fields;
    fieldOrder[entry.type] = parsed.order;
  }

  for (const staleDocId of knownEmptyUnused) {
    problems.push({
      type: null,
      docId: staleDocId,
      reason: 'listed in KNOWN_NO_PROPERTIES but now has a "### Properties" table (or no longer exists) - update this script',
    });
  }

  return { registryCount: registry.length, schema: result, fieldOrder, problems };
}

function main() {
  const { registryCount, schema, problems } = extractAll();
  const parsedCount = Object.keys(schema).length - KNOWN_NO_PROPERTIES.size;

  // The human-readable report always goes to stderr so `--json` output on stdout stays pure
  // JSON and pipeable (e.g. `node extract-segment-properties.mjs --json > schema.json`).
  const report = console.error;

  report(`Parsed ${parsedCount}/${registryCount - KNOWN_NO_PROPERTIES.size} segment docs' ` +
    `"### Properties" tables (${problems.length} unparsed, ` +
    `${KNOWN_NO_PROPERTIES.size} known to legitimately have none: ${[...KNOWN_NO_PROPERTIES].join(', ')}).`);

  if (problems.length > 0) {
    report(`\n${problems.length} doc(s) could not be parsed:`);
    for (const p of problems) {
      report(`  - ${p.type ?? '(n/a)'} (docId: ${p.docId}): ${p.reason}`);
    }
  } else {
    report('No parse problems.');
  }

  if (process.argv.includes('--json')) {
    console.log(JSON.stringify(schema, null, 2));
  }

  process.exitCode = problems.length > 0 ? 1 : 0;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
