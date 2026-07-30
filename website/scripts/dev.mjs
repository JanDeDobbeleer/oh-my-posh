#!/usr/bin/env node
// `npm start` runs this instead of `docusaurus start` directly, so local edits are reflected
// live without restarting the dev server:
//
//  - Editing/adding/removing a file under /themes re-renders *only that theme* and patches it
//    into generated/themes.json - the file <ThemeGallery/> statically imports, so webpack's own
//    dev-server watcher (already watching every file in the module graph) picks up the rewrite
//    and hot-reloads the gallery page. A full `npm run themes` re-renders all 124+ themes, one
//    child-process spawn each; patching just the changed entry turns a multi-second wait after
//    every save into a single render.
//  - Editing export_themes.mjs, font-metrics.mjs or segment_data.json - anything that can change
//    every theme's render, not just one - falls back to a full `npm run themes`.
//  - Editing a Go source file under src/ rebuilds the website's generated artifacts (`node
//    scripts/ensure-artifacts.mjs`), so a renderer change shows up in both the theme gallery and
//    the in-browser studio without restarting anything.
//
// None of this runs for `npm run build` or CI: both already run `npm run themes` / `npm run
// wasm` explicitly before building (see ensure-artifacts.mjs's own comment on that split), and
// neither watches anything.
//
// A Node script rather than a second npm script run via a process-concurrency package: this
// repo's developers run PowerShell, cmd and POSIX shells (see ensure-artifacts.mjs's own comment
// on the same point), and spawning/forwarding signals by hand here avoids a new dependency for
// what is otherwise a few lines of child_process.
import { spawn, spawnSync } from 'node:child_process';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { createServer } from 'node:net';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import chokidar from 'chokidar';

import { CONFIG, exportTheme, getThemeNameFromFile, isValidTheme } from '../export_themes.mjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const REPO_DIR = join(WEBSITE_DIR, '..');
const SRC_DIR = join(REPO_DIR, 'src');
const IS_WINDOWS = process.platform === 'win32';

// usePolling mirrors `docusaurus start --poll` below: some developers run this repo from
// filesystems (network drives, some container/VM setups) where native OS file-change events
// never fire, and every watcher here needs to behave the same way or only some of them would
// work there.
const POLLING_OPTIONS = { usePolling: true, interval: 1000 };

async function readManifest() {
  try {
    return JSON.parse(await readFile(CONFIG.MANIFEST_FILE, 'utf8'));
  } catch {
    // Missing or malformed: ensure-artifacts.mjs (prestart) already ran `npm run themes` once,
    // so this only happens if that file was deleted mid-session. Starting from empty is fine -
    // the next full regen or per-theme reload rebuilds it.
    return [];
  }
}

async function writeManifest(manifest) {
  await mkdir(dirname(CONFIG.MANIFEST_FILE), { recursive: true });
  await writeFile(CONFIG.MANIFEST_FILE, JSON.stringify(manifest));
}

// Reruns exactly one theme's render and splices the result into the existing manifest, in place
// when the theme already has an entry (the common case: an edit) or appended and re-sorted when
// it doesn't (a new theme file). Sorting only applies to inserts, so this can drift from
// export_themes.mjs's own filename-sorted order over a long dev session of adds/renames - a full
// `npm run themes` (which every commit's CI already runs) is the source of truth for that order.
async function reloadTheme(themeFile) {
  console.log(`[themes] ${themeFile} changed, re-rendering...`);

  let result;

  try {
    result = await exportTheme(themeFile);
  } catch (error) {
    console.error(`[themes] Failed to render ${themeFile}: ${error.message}`);
    return;
  }

  if (!result) {
    return;
  }

  const manifest = await readManifest();
  const index = manifest.findIndex((entry) => entry.name === result.entry.name);

  if (index === -1) {
    manifest.push(result.entry);
    manifest.sort((a, b) => a.name.localeCompare(b.name));
  } else {
    manifest[index] = result.entry;
  }

  await writeManifest(manifest);
  console.log(`[themes] Updated "${result.entry.name}" in the gallery.`);
}

async function removeTheme(themeFile) {
  const themeName = getThemeNameFromFile(themeFile);
  const manifest = await readManifest();
  const next = manifest.filter((entry) => entry.name !== themeName);

  if (next.length === manifest.length) {
    return;
  }

  await writeManifest(next);
  console.log(`[themes] Removed "${themeName}" from the gallery (file deleted).`);
}

// cmd.exe splits on whitespace, so anything containing a space or quote needs quoting; doubling
// embedded quotes is cmd's own escape convention.
function quoteForCmd(arg) {
  return /[\s"]/.test(arg) ? `"${arg.replace(/"/g, '""')}"` : arg;
}

function runNpmScript(script) {
  // Node deprecates passing an args array alongside shell: true (DEP0190) - it can't escape them
  // for an unknown shell, so on Windows the full, already-quoted command is built as one string
  // instead and given no separate args.
  const [command, args] = IS_WINDOWS ? [['npm', 'run', script].map(quoteForCmd).join(' '), []] : ['npm', ['run', script]];
  const result = spawnSync(command, args, { cwd: WEBSITE_DIR, stdio: 'inherit', shell: IS_WINDOWS });

  if (result.status === 0) {
    return true;
  }

  console.error(`[dev] npm run ${script} exited with code ${result.status}. Fix the error above and save again.\n`);
  return false;
}

function fullThemesRegen() {
  console.log('\n[themes] export_themes.mjs / font-metrics.mjs / segment_data.json changed, regenerating the full gallery...');

  if (runNpmScript('themes')) {
    console.log('[themes] Theme gallery regenerated.\n');
  }
}

function rebuildArtifacts() {
  console.log('\n[artifacts] Go source changed, rebuilding website artifacts...');

  const result = spawnSync(process.execPath, [join(__dirname, 'ensure-artifacts.mjs')], {
    cwd: WEBSITE_DIR,
    stdio: 'inherit',
  });

  if (result.status === 0) {
    console.log('[artifacts] Website artifacts rebuilt.\n');
    return true;
  }

  console.error(`[artifacts] node scripts/ensure-artifacts.mjs exited with code ${result.status}. Fix the error above and save again.\n`);
  return false;
}

// A tiny per-key debounce/serialize queue: coalesces rapid-fire events for the same key (a save
// that touches a file twice, or an editor writing a temp file then renaming it over the
// original) into one run, and never overlaps two runs for the same key.
function createDebouncedRunner(delayMs) {
  const timers = new Map();
  const running = new Set();
  const rerun = new Map();

  function execute(key, fn) {
    if (running.has(key)) {
      rerun.set(key, fn);
      return;
    }

    running.add(key);

    Promise.resolve()
      .then(fn)
      .catch((error) => console.error(`[dev] ${error.message}`))
      .finally(() => {
        running.delete(key);

        const next = rerun.get(key);

        if (next) {
          rerun.delete(key);
          execute(key, next);
        }
      });
  }

  return function schedule(key, fn) {
    clearTimeout(timers.get(key));
    timers.set(key, setTimeout(() => execute(key, fn), delayMs));
  };
}

const scheduleThemeWork = createDebouncedRunner(300);
const scheduleSharedWork = createDebouncedRunner(500);
const scheduleArtifactWork = createDebouncedRunner(500);

const themesWatcher = chokidar.watch(CONFIG.THEMES_CONFIG_DIR, { ignoreInitial: true, ...POLLING_OPTIONS });

themesWatcher.on('all', (event, path) => {
  const themeFile = path.split(/[/\\]/).pop();

  if (!isValidTheme(themeFile)) {
    return;
  }

  if (event === 'unlink') {
    scheduleThemeWork(themeFile, () => removeTheme(themeFile));
  } else if (event === 'add' || event === 'change') {
    scheduleThemeWork(themeFile, () => reloadTheme(themeFile));
  }
});

const sharedWatcher = chokidar.watch(
  [join(WEBSITE_DIR, 'export_themes.mjs'), join(WEBSITE_DIR, 'font-metrics.mjs'), CONFIG.SEGMENT_DATA_FILE],
  { ignoreInitial: true, ...POLLING_OPTIONS },
);

sharedWatcher.on('all', () => scheduleSharedWork('themes', fullThemesRegen));

const wasmWatcher = chokidar.watch([join(SRC_DIR, '**/*.go'), join(SRC_DIR, 'go.mod'), join(SRC_DIR, 'go.sum')], {
  ignored: '**/*_test.go',
  ignoreInitial: true,
  ...POLLING_OPTIONS,
});

wasmWatcher.on('all', () => scheduleArtifactWork('artifacts', rebuildArtifacts));

console.log('[dev] Watching /themes (per-theme reload), the exporter/data files (full regen) and src/**/*.go (artifact rebuild)...');

async function ensurePortAvailable(port) {
  const server = createServer();

  return new Promise((resolve, reject) => {
    server.once('error', (error) => {
      if (error.code === 'EADDRINUSE') {
        reject(new Error(`Port ${port} is already in use. Stop the existing dev server or set PORT to a different value.`));
        return;
      }

      reject(error);
    });

    server.once('listening', () => {
      server.close(() => resolve());
    });

    server.listen(port, 'localhost');
  });
}

const port = process.env.PORT || '3000';

try {
  await ensurePortAvailable(port);
} catch (error) {
  console.error(error.message);
  process.exit(1);
}

const docusaurusBin = join(WEBSITE_DIR, 'node_modules', '.bin', IS_WINDOWS ? 'docusaurus.cmd' : 'docusaurus');
const docusaurusArgs = ['start', '--poll', '1000', '--port', port];
// Same DEP0190 workaround as runNpmScript above: shell: true and an args array don't mix.
const [docusaurusCommand, spawnArgs] = IS_WINDOWS
  ? [[docusaurusBin, ...docusaurusArgs].map(quoteForCmd).join(' '), []]
  : [docusaurusBin, docusaurusArgs];
const docusaurus = spawn(docusaurusCommand, spawnArgs, {
  cwd: WEBSITE_DIR,
  stdio: 'inherit',
  shell: IS_WINDOWS,
});

let shuttingDown = false;

function shutdown(signal) {
  if (shuttingDown) {
    return;
  }

  shuttingDown = true;
  themesWatcher.close();
  sharedWatcher.close();
  wasmWatcher.close();
  docusaurus.kill(signal);
}

process.on('SIGINT', () => shutdown('SIGINT'));
process.on('SIGTERM', () => shutdown('SIGTERM'));

docusaurus.on('exit', (code, signal) => {
  themesWatcher.close();
  sharedWatcher.close();
  wasmWatcher.close();

  if (shuttingDown) {
    return;
  }

  process.exit(signal ? 1 : code ?? 0);
});
