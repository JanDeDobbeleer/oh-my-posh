import { exec } from 'node:child_process';
import { promises } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

const execAsync = promisify(exec);
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const CONFIG = {
  THEMES_CONFIG_DIR: join(__dirname, '../themes'),
  THEMES_STATIC_DIR: join(__dirname, 'static/img/themes'),
  OUTPUT_FILE: join(__dirname, 'docs/themes.md'),
  CONCURRENCY: 8,
  DEFAULT_BG_COLOR: '#151515',
  THEME_EXTENSIONS: ['.omp.json', '.omp.toml', '.omp.yaml'],
  SEGMENT_DATA_FILE: join(__dirname, 'segment_data.json'),
  GITHUB_BASE_URL: 'https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes'
};

const TRENDING_FETCH_TIMEOUT_MS = 4000;

/**
 * Small, tasteful deny-list used as a safety net behind each source's own
 * explicit-content flag. Matched word-boundary aware so substrings inside
 * unrelated words (e.g. "sex" in "Essex") don't trigger a false positive.
 */
const CONTENT_DENY_LIST = ['fuck', 'shit', 'bitch', 'nigga', 'cunt', 'motherfucker'];
const CONTENT_DENY_REGEX = new RegExp(`\\b(${CONTENT_DENY_LIST.join('|')})\\b`, 'i');

const THEME_CONFIG_OVERRIDES = new Map([
  ['amro.omp.json', { author: 'AmRo', bgColor: '#1C2029' }],
  ['chips.omp.json', {
    author: 'CodexLink | v1.2.4, Single Width (07/11/2023) | https://github.com/CodexLink/chips.omp.json',
    bgColor: CONFIG.DEFAULT_BG_COLOR
  }],
  ['craver.omp.json', { author: 'Nick Craver', bgColor: '#282c34' }],
  ['hunk.omp.json', { author: 'Paris Qian', bgColor: CONFIG.DEFAULT_BG_COLOR }],
  ['kushal.omp.json', { author: 'Kushal-Chandar', bgColor: CONFIG.DEFAULT_BG_COLOR }],
  ['night-owl.omp.json', { author: 'Mr-Vipi', bgColor: '#011627' }],
  ['quick-term.omp.json', { author: 'SokLay', bgColor: CONFIG.DEFAULT_BG_COLOR }],
  ['catppuccin.omp.json', { author: 'IrwinJuice', bgColor: '#24273A' }],
  ['catppuccin_latte.omp.json', { author: 'IrwinJuice', bgColor: '#EFF1F5' }],
  ['catppuccin_frappe.omp.json', { author: 'IrwinJuice', bgColor: '#303446' }],
  ['catppuccin_macchiato.omp.json', { author: 'IrwinJuice', bgColor: '#24273A' }],
  ['catppuccin_mocha.omp.json', { author: 'IrwinJuice', bgColor: '#1E1E2E' }]
]);

function createThemeConfig(author = '', bgColor = CONFIG.DEFAULT_BG_COLOR) {
  return { author, bgColor };
}

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

function buildPoshCommand(configPath, outputImage, config) {
  const parts = [
    'oh-my-posh config export image',
    `--config=${configPath}`,
    `--output=${outputImage}`,
    `--background-color=${config.bgColor}`,
    `--data=${CONFIG.SEGMENT_DATA_FILE}`,
  ];

  if (config.author) {
    parts.push(`--author="${config.author}"`);
  }

  return parts.join(' ');
}

function generateThemeMarkdown(themeName, themeFile) {
  const themeData = `
### [${themeName}]

[![${themeName}](/img/themes/${themeName}.png)][${themeName}]
`;

  const link = `[${themeName}]: ${CONFIG.GITHUB_BASE_URL}/${themeFile} '${themeName}'\n`;

  return { themeData, link };
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

  try {
    const configPath = join(CONFIG.THEMES_CONFIG_DIR, themeFile);
    const config = THEME_CONFIG_OVERRIDES.get(themeFile) || createThemeConfig();
    const themeName = getThemeNameFromFile(themeFile);
    const imageFile = `${themeName}.png`;
    const outputPath = join(CONFIG.THEMES_STATIC_DIR, imageFile);

    const poshCommand = buildPoshCommand(configPath, outputPath, config);
    const { stderr } = await execAsync(poshCommand);

    if (stderr) {
      console.error(`Unable to create image for ${themeFile}: ${stderr}`);
      return null;
    }

    console.info(`Exported ${themeFile} to ${outputPath}`);

    const { themeData, link } = generateThemeMarkdown(themeName, themeFile);

    return { themeData, link, fileName: themeFile };

  } catch (error) {
    console.error(`Error processing theme ${themeFile}:`, error.message);
    return null;
  }
}

async function ensureDirectories() {
  try {
    await promises.access(CONFIG.THEMES_STATIC_DIR);
  } catch {
    await promises.mkdir(CONFIG.THEMES_STATIC_DIR, { recursive: true });
  }
}

async function main() {
  try {
    console.log('Starting theme export process...');

    await ensureDirectories();

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

    // Sort by filename keys alphabetically
    const sortedFileNames = Array.from(resultsMap.keys()).sort();

    // Append theme data to the file in sorted order
    for (const fileName of sortedFileNames) {
      const result = resultsMap.get(fileName);
      await promises.appendFile(CONFIG.OUTPUT_FILE, result.themeData);
    }

    // Add separator line
    await promises.appendFile(CONFIG.OUTPUT_FILE, '\n');

    // Append all links in the same sorted order
    for (const fileName of sortedFileNames) {
      const result = resultsMap.get(fileName);
      await promises.appendFile(CONFIG.OUTPUT_FILE, result.link);
    }

    console.log(`Successfully exported ${resultsMap.size} themes to ${CONFIG.OUTPUT_FILE}`);

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
  createThemeConfig,
  isValidTheme,
  getThemeNameFromFile,
  generateThemeMarkdown,
  asyncPool,
  main,
  fetchJsonWithTimeout,
  trendingFromDeezer,
  trendingFromAppleRSS,
  isClean,
  buildDataFileWithTrending,
};
