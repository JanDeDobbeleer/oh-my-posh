#!/usr/bin/env node
// Renders static/img/og-image.png: the social card link previews show for ohmyposh.dev.
//
// It is a raster on purpose - Twitter, Slack, Discord, iMessage and the rest will not render an
// SVG - so this is the one image on the site that cannot just be the inlined prompt. It composes
// the same rendered default prompt the homepage shows (generated/hero.json, see
// export_themes.mjs) onto a 1200x630 card and screenshots it through headless Chrome, which is
// what gets the real Victor Mono glyphs: an SVG drawn into a canvas cannot see the page's
// @font-face, so the icons and powerline separators would fall back to a system font.
//
// NOT part of the build. Chrome is not a dependency the docs build should grow, and the card only
// changes when the default config or the font does, which is rare. The output is committed like
// any other static asset. Re-run by hand after either changes:
//
//   node scripts/render-og-image.mjs
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const HERO = join(WEBSITE_DIR, 'generated', 'hero.json');
const FONT = join(WEBSITE_DIR, 'static', 'fonts', 'VictorMono.ttf');
const OUT = join(WEBSITE_DIR, 'static', 'img', 'og-image.png');

// The canonical Open Graph size. Every consumer crops or letterboxes to its own ratio from this,
// and 1200x630 is what Facebook/Twitter/LinkedIn all document as the safe one.
const WIDTH = 1200;
const HEIGHT = 630;

const CHROME_CANDIDATES = [
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  '/usr/bin/google-chrome',
  '/usr/bin/chromium',
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
];

function findChrome() {
  const found = CHROME_CANDIDATES.find((candidate) => existsSync(candidate));

  if (!found) {
    throw new Error(
      'No Chrome or Edge found. This script screenshots a page to get the real font; ' +
      `looked in:\n  ${CHROME_CANDIDATES.join('\n  ')}`,
    );
  }

  return found;
}

function buildHTML(svg, fontBase64) {
  // Colours mirror the site's own dark theme (src/css/custom.css) so the card reads as part of
  // the same surface a visitor lands on. The prompt keeps its own terminal window chrome, which
  // is what makes the card legible at thumbnail size: a bordered dark rectangle on a dark field.
  return `<!doctype html>
<meta charset="utf-8">
<style>
  @font-face {
    font-family: "Victor Mono";
    src: url(data:font/ttf;base64,${fontBase64}) format("truetype");
  }

  html, body { margin: 0; padding: 0; }

  body {
    width: ${WIDTH}px;
    height: ${HEIGHT}px;
    background: #1b1b1d;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 44px;
    font-family: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
    overflow: hidden;
  }

  h1 {
    margin: 0;
    font-size: 76px;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: #ffffff;
  }

  p {
    margin: -28px 0 0;
    font-size: 28px;
    color: #b8bcc4;
  }

  .prompt { width: 1020px; }
  .prompt svg { display: block; width: 100%; height: auto; }
</style>
<h1>Oh My Posh</h1>
<p>The most customizable and fastest prompt engine for any shell.</p>
<div class="prompt">${svg}</div>
`;
}

const hero = JSON.parse(readFileSync(HERO, 'utf8'));

if (typeof hero.svg !== 'string' || !hero.svg) {
  throw new Error(`${HERO} has no svg. Run \`npm run themes\` first.`);
}

const chrome = findChrome();
const work = mkdtempSync(join(tmpdir(), 'omp-og-'));
const page = join(work, 'card.html');

try {
  writeFileSync(page, buildHTML(hero.svg, readFileSync(FONT).toString('base64')));

  execFileSync(chrome, [
    '--headless',
    '--disable-gpu',
    '--hide-scrollbars',
    // Chrome writes screenshot.png into the working directory it is given, not to an arbitrary
    // path, so it runs in the scratch directory and the result is moved afterwards.
    `--screenshot=${join(work, 'shot.png')}`,
    `--window-size=${WIDTH},${HEIGHT}`,
    // The font is a 2.4MB data URI; the default virtual time budget can expire before it is
    // parsed and applied, which silently produces a card in a fallback font.
    '--virtual-time-budget=10000',
    `file://${page.replace(/\\/g, '/')}`,
  ], { stdio: 'pipe' });

  const shot = join(work, 'shot.png');

  if (!existsSync(shot)) {
    throw new Error('Chrome exited without writing a screenshot.');
  }

  writeFileSync(OUT, readFileSync(shot));

  const { size } = statSync(OUT);
  console.log(`Wrote ${OUT} (${WIDTH}x${HEIGHT}, ${(size / 1024).toFixed(0)} KB)`);
} finally {
  rmSync(work, { recursive: true, force: true });
}
