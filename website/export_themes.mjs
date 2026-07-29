import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promises } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

import { VICTOR_MONO } from './font-metrics.mjs';

const execFileAsync = promisify(execFile);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// VICTOR_MONO (family name plus CELL_WIDTH/LINE_HEIGHT/FILL_ASCENT/FILL_DESCENT) lives in
// font-metrics.mjs now, not here - that file has no Node imports, so it can also be imported by
// the studio's browser-side code (website/src/components/Studio), which cannot import this
// script (it pulls in node:fs et al.). See font-metrics.mjs for the derivation comments; keep
// them there, not duplicated here.

const CONFIG = {
  THEMES_CONFIG_DIR: join(__dirname, '../themes'),
  // Read by plugins/themes at build time (usePluginData('oh-my-posh-themes')) and
  // inlined into docs/themes.mdx via <ThemeGallery/>. Not under static/: nothing
  // references these SVGs by URL any more (they're inlined, not <img>-loaded), so
  // living outside static/ keeps them from also being copied into the deployed
  // build/ output as dead weight. Regenerated on every build (see the "themes" npm
  // script) and gitignored, exactly like the PNGs it replaces.
  MANIFEST_FILE: join(__dirname, 'generated/themes.json'),
  CONCURRENCY: 8,
  THEME_EXTENSIONS: ['.omp.json', '.omp.toml', '.omp.yaml'],
  SEGMENT_DATA_FILE: join(__dirname, 'segment_data.json'),
  GITHUB_BASE_URL: 'https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes',
  FONT_FAMILY: VICTOR_MONO.FONT_FAMILY,
  CELL_WIDTH: VICTOR_MONO.CELL_WIDTH,
  LINE_HEIGHT: VICTOR_MONO.LINE_HEIGHT,
  FILL_ASCENT: VICTOR_MONO.FILL_ASCENT,
  FILL_DESCENT: VICTOR_MONO.FILL_DESCENT,
};

const TRENDING_FETCH_TIMEOUT_MS = 4000;

/**
 * Small, tasteful deny-list used as a safety net behind each source's own
 * explicit-content flag. Matched word-boundary aware so substrings inside
 * unrelated words (e.g. "sex" in "Essex") don't trigger a false positive.
 */
const CONTENT_DENY_LIST = ['fuck', 'shit', 'bitch', 'nigga', 'cunt', 'motherfucker'];
const CONTENT_DENY_REGEX = new RegExp(`\\b(${CONTENT_DENY_LIST.join('|')})\\b`, 'i');

// There used to be a THEME_CONFIG_OVERRIDES map here (per-theme author/bgColor
// overrides). Both only ever fed the PNG path's --author/--background-color
// flags - image.Settings.Author's caption and image.Renderer's canvas fill -
// neither of which the svg format reads (see config_export_svg.go: it takes
// the terminal background from the theme's own config, not a CLI flag). Now
// that every theme exports as svg, the overrides had no effect left to have.

function isValidTheme(fileName) {
  return CONFIG.THEME_EXTENSIONS.some((ext) => fileName.endsWith(ext));
}

function getThemeNameFromFile(fileName) {
  const lastDotIndex = fileName.lastIndexOf('.');
  const secondLastDotIndex = fileName.lastIndexOf('.', lastDotIndex - 1);
  return fileName.slice(0, secondLastDotIndex);
}

async function fetchJsonWithTimeout(url) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), TRENDING_FETCH_TIMEOUT_MS);

  try {
    const response = await fetch(url, { signal: controller.signal });

    if (!response.ok) {
      throw new Error(`request failed with status ${response.status}`);
    }

    return await response.json();
  } finally {
    clearTimeout(timeout);
  }
}

/**
 * Plain-text safety net behind each source's own explicit flag - intentionally
 * short rather than a comprehensive filter.
 */
function isClean(title, artist) {
  return !CONTENT_DENY_REGEX.test(`${title} ${artist}`);
}

async function trendingFromDeezer() {
  const body = await fetchJsonWithTimeout('https://api.deezer.com/chart/0/tracks?limit=25');
  const tracks = body?.data;

  if (!Array.isArray(tracks) || tracks.length === 0) {
    throw new Error('no tracks returned');
  }

  const track = tracks.find((entry) => !entry.explicit_lyrics && isClean(entry.title, entry.artist?.name));

  if (!track) {
    throw new Error('no clean track found in chart');
  }

  return { artist: track.artist.name, track: track.title };
}

async function trendingFromAppleRSS() {
  const body = await fetchJsonWithTimeout('https://rss.marketingtools.apple.com/api/v2/us/music/most-played/25/songs.json');
  const tracks = body?.feed?.results;

  if (!Array.isArray(tracks) || tracks.length === 0) {
    throw new Error('no tracks returned');
  }

  // Apple only includes contentAdvisoryRating when a track is explicit, so presence of the key
  // (not its value) is the signal to filter on. The value is the literal string "Explict" (sic),
  // Apple's own misspelling, verified against the live feed - do not "correct" it here.
  const track = tracks.find((entry) => !('contentAdvisoryRating' in entry) && isClean(entry.name, entry.artistName));

  if (!track) {
    throw new Error('no clean track found in feed');
  }

  return { artist: track.artistName, track: track.name };
}

/**
 * Fetches a trending track (Deezer, then Apple Music RSS as a fallback) and injects
 * it into the spotify and ytm segment payloads of a temporary copy of the committed
 * data file. The committed file is never modified; on any failure the committed
 * file is used as-is.
 */
async function buildDataFileWithTrending() {
  let trending;

  try {
    trending = await trendingFromDeezer();
  } catch (error) {
    console.warn(`Trending track lookup via Deezer failed: ${error.message}`);

    try {
      trending = await trendingFromAppleRSS();
    } catch (fallbackError) {
      console.warn(`Trending track lookup via Apple Music RSS failed: ${fallbackError.message}`);
    }
  }

  if (!trending) {
    return CONFIG.SEGMENT_DATA_FILE;
  }

  try {
    const raw = await promises.readFile(CONFIG.SEGMENT_DATA_FILE, 'utf8');
    const data = JSON.parse(raw);

    for (const key of ['spotify', 'ytm']) {
      const segment = data.segments?.[key];

      if (!segment) {
        continue;
      }

      segment.Artist = trending.artist;
      segment.Track = trending.track;
    }

    const tempPath = join(tmpdir(), `segment_data.${process.pid}.${Date.now()}.json`);
    await promises.writeFile(tempPath, JSON.stringify(data, null, 2));

    console.log(`Using trending track "${trending.track}" by ${trending.artist} for previews`);

    return tempPath;
  } catch (error) {
    console.warn(`Unable to build data file with trending track: ${error.message}`);
    return CONFIG.SEGMENT_DATA_FILE;
  }
}

// OMP_BIN points the export at a specific oh-my-posh binary instead of
// whatever is on PATH, so a gallery can be regenerated from a local build
// without installing it over the one the developer actually uses. CI and the
// normal build leave it unset and get PATH, as before. execFile spawns the
// binary directly with an argv array (no shell in between), so - unlike the
// exec()+string command this replaced - the path never needs quoting even
// when it contains spaces.
const OMP_BIN = process.env.OMP_BIN || 'oh-my-posh';

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
    // segment_data.json is hand-written on purpose (see buildDataFileWithTrending):
    // its synthetic values are what make the renders look like a plausible
    // machine. oh-my-posh warns that the file carries no recorder marker, which is
    // expected here and handled as a non-fatal stderr line below. --data-derive
    // would not silence it - that flag only forces an already-recorded file to
    // re-derive, so it is a no-op for a hand-written one.
    `--data=${CONFIG.SEGMENT_DATA_FILE}`,
    // Without this the build machine leaks into the gallery. A segment the data
    // file does not cover used to derive itself from whatever machine ran the
    // export: free-ukraine and markbull published a worktree count read from the
    // exporting checkout's own .git/worktrees, so the number changed with whoever
    // built the site. --data-only cuts the environment off, leaving a segment
    // either rendering from this file or reporting itself absent - never from the
    // machine. It does not suppress segments the file misses: path and session own
    // no entry and still render from the env section's pinned PWD and user.
    '--data-only',
  ];
}

// buildManifestEntry builds the one object plugins/themes hands to
// <ThemeGallery/> per theme: the raw svg markup (inlined verbatim via
// dangerouslySetInnerHTML, never parsed as MDX/JSX - MDX compiles a document's body as JSX, and
// the exporter's raw SVG attributes are not valid JSX, so the markup has to arrive as an opaque
// string rather than live in the .mdx source itself), the name to label it
// with, and the GitHub URL both the heading and the render link to.
function buildManifestEntry(themeName, themeFile, svg) {
  return {
    name: themeName,
    githubUrl: `${CONFIG.GITHUB_BASE_URL}/${themeFile}`,
    svg,
  };
}

async function* asyncPool(concurrency, iterable, iteratorFn) {
  const executing = new Set();

  async function consume() {
    const [promise, value] = await Promise.race(executing);
    executing.delete(promise);
    return value;
  }

  for (const item of iterable) {
    const promise = (async () => await iteratorFn(item))().then(
      value => [promise, value]
    );
    executing.add(promise);

    if (executing.size >= concurrency) {
      yield await consume();
    }
  }

  while (executing.size) {
    yield await consume();
  }
}

async function exportTheme(themeFile) {
  if (!isValidTheme(themeFile)) {
    return null;
  }

  const configPath = join(CONFIG.THEMES_CONFIG_DIR, themeFile);
  const themeName = getThemeNameFromFile(themeFile);
  // A per-run scratch path, like buildDataFileWithTrending's temp data file below:
  // the svg only needs to exist long enough to be read back into the manifest, so
  // it lives in the OS temp dir rather than under website/, leaving nothing behind
  // for git status to notice.
  const outputPath = join(tmpdir(), `omp-theme-${randomUUID()}.svg`);
  const poshArgs = buildPoshArgs(configPath, outputPath);

  let stderr;

  try {
    ({ stderr } = await execFileAsync(OMP_BIN, poshArgs));
  } catch (error) {
    // execFileAsync only rejects on a non-zero exit code - a genuine render failure,
    // not incidental stderr output. Fail the build loudly instead of silently
    // dropping the theme from the page.
    throw new Error(`Failed to export theme ${themeFile}: ${error.message}`);
  }

  if (stderr) {
    // A non-zero exit already threw above, so stderr here is incidental (e.g. the
    // hand-written-data-file warning) and must not fail the build.
    console.warn(`${themeFile}: ${stderr.trim()}`);
  }

  let svg;

  try {
    svg = await promises.readFile(outputPath, 'utf8');
  } finally {
    await promises.unlink(outputPath).catch(() => {});
  }

  console.info(`Exported ${themeFile}`);

  return { entry: buildManifestEntry(themeName, themeFile, svg), fileName: themeFile };
}

async function ensureDirectories() {
  // recursive: true is idempotent (no error if the directory already exists), so there is nothing
  // an access()-then-mkdir() dance would catch that a bare mkdir() doesn't already handle.
  await promises.mkdir(dirname(CONFIG.MANIFEST_FILE), { recursive: true });
}

async function main() {
  try {
    console.log('Starting theme export process...');

    await ensureDirectories();

    const committedDataFile = CONFIG.SEGMENT_DATA_FILE;
    CONFIG.SEGMENT_DATA_FILE = await buildDataFileWithTrending();

    const themes = await promises.readdir(CONFIG.THEMES_CONFIG_DIR);
    const validThemes = themes.filter(isValidTheme);

    console.log(`Found ${validThemes.length} valid themes to process`);

    const resultsMap = new Map();

    for await (const result of asyncPool(CONFIG.CONCURRENCY, validThemes, exportTheme)) {
      if (result) {
        // Use the original filename as the key for efficient sorting
        resultsMap.set(result.fileName, result);
      }
    }

    // Sort by filename keys alphabetically - the same order the page has shown
    // since the PNG era, preserved here even though the sort key (the source
    // config's filename) is no longer part of the emitted manifest entry itself.
    const sortedFileNames = Array.from(resultsMap.keys()).sort();
    const manifest = sortedFileNames.map((fileName) => resultsMap.get(fileName).entry);

    await promises.writeFile(CONFIG.MANIFEST_FILE, JSON.stringify(manifest));

    console.log(`Successfully exported ${manifest.length} themes to ${CONFIG.MANIFEST_FILE}`);

    // buildDataFileWithTrending only returns a path other than the committed file when it
    // actually wrote a scratch copy with the trending track baked in (see its own comment); on any
    // failure - fetch failed, write failed - it falls back to returning the committed file itself.
    // Only clean up the scratch copy: the committed file must never be deleted, and skipping this
    // check would do exactly that on every run where trending lookup failed.
    if (CONFIG.SEGMENT_DATA_FILE !== committedDataFile) {
      await promises.unlink(CONFIG.SEGMENT_DATA_FILE).catch(() => {});
    }

  } catch (error) {
    console.error('Export process failed:', error.message);
    process.exit(1);
  }
}

// Execute main function if this file is run directly
// In ES modules, we check if import.meta.url matches the process argv
if (process.argv[1] === __filename) {
  main();
}

export {
  exportTheme,
  isValidTheme,
  getThemeNameFromFile,
  buildManifestEntry,
  asyncPool,
  main,
  fetchJsonWithTimeout,
  trendingFromDeezer,
  trendingFromAppleRSS,
  isClean,
  buildDataFileWithTrending,
};
