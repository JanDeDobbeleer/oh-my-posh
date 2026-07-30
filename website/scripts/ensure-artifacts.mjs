#!/usr/bin/env node
// Generates the three build artifacts the site cannot start without, but only the ones that are
// actually missing or out of date. Wired as prestart/prebuild (see package.json), so `npm run
// start` from a fresh clone works instead of failing in the themes plugin's loadContent with a
// note about a file it cannot write itself.
//
// CI does not need this - .github/workflows/docs.yml still runs the three scripts as explicit
// steps before building, and by the time the Docusaurus build runs every artifact is present and
// newer than its sources, so every check below short-circuits.
//
// A Node script rather than shell: it has to build a Go binary, probe it, compare mtimes and
// spawn npm, and this repo's developers run PowerShell, cmd and POSIX shells (see
// build-wasm.mjs's own comment on the same point).
import { execFileSync, spawnSync } from 'node:child_process';
import { existsSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const REPO_DIR = join(WEBSITE_DIR, '..');
const SRC_DIR = join(REPO_DIR, 'src');
const BIN = join(SRC_DIR, 'bin', process.platform === 'win32' ? 'oh-my-posh.exe' : 'oh-my-posh');

// The flag every one of these renders passes. It is also the newest of them, so a binary that
// knows it is new enough to have the rest - which is exactly the failure a stale oh-my-posh on
// PATH produces: "unknown flag: --data-only", 117 times.
const REQUIRED_FLAG = '--data-only';

// Each artifact, the npm script that writes it, and the inputs that invalidate it. Go sources
// count for all three: the renderer decides what every one of them contains.
const ARTIFACTS = [
  {
    name: 'theme gallery',
    script: 'themes',
    // Two outputs from one script: the full manifest the gallery imports, and the single default
    // theme the homepage does (see export_themes.mjs's HERO_FILE). Both have to be checked, or
    // adding the hero to an otherwise up-to-date checkout would never regenerate.
    output: [join(WEBSITE_DIR, 'generated', 'themes.json'), join(WEBSITE_DIR, 'generated', 'hero.json')],
    inputs: [SRC_DIR, ...binaryInput(), join(REPO_DIR, 'themes'), join(WEBSITE_DIR, 'export_themes.mjs'), join(WEBSITE_DIR, 'segment_data.json')],
    needsCLI: true,
  },
  {
    name: 'segment previews',
    script: 'segment-previews',
    output: [join(WEBSITE_DIR, 'generated', 'segment-previews.json')],
    inputs: [SRC_DIR, ...binaryInput(), join(WEBSITE_DIR, 'scripts', 'render-segment-previews.mjs'), join(WEBSITE_DIR, 'segment_data.json'), join(WEBSITE_DIR, 'docs', 'segments')],
    needsCLI: true,
  },
  {
    name: 'studio wasm module',
    script: 'wasm',
    output: [join(WEBSITE_DIR, 'static', 'omp.wasm')],
    inputs: [SRC_DIR],
    needsCLI: false,
  },
];

// Directories that hold no input to any render but plenty of files, so walking them only costs
// time. bin/ is walked past here and handled separately by binaryInput below, which cares about
// exactly one file in it.
const SKIP_DIRS = new Set(['bin', 'node_modules', '.git', 'testdata', 'generated']);

// The binary that renders counts as an input, not just the Go sources behind it. Comparing only
// against src/ says an artifact is current whenever it was written after the last source edit -
// but a render is only as new as the oh-my-posh that produced it, and rebuilding the binary
// touches nothing under src/. That gap shipped a homepage rendered by a build from before two
// renderer fixes: sources 11:52, artifact 12:54, binary 13:06, and the check saw 12:54 > 11:52
// and called it fresh.
//
// Only a real file can be compared. OMP_BIN naming something on PATH, or no local build at all,
// leaves this out and the source mtimes stand on their own.
function binaryInput() {
  const candidate = process.env.OMP_BIN && process.env.OMP_BIN.includes(sep)
    ? process.env.OMP_BIN
    : BIN;

  return existsSync(candidate) ? [candidate] : [];
}

function newestMtime(path) {
  let newest = 0;

  const visit = (current) => {
    let info;

    try {
      info = statSync(current);
    } catch {
      // A listed input that does not exist yet cannot invalidate anything.
      return;
    }

    if (!info.isDirectory()) {
      newest = Math.max(newest, info.mtimeMs);
      return;
    }

    for (const entry of readdirSync(current, { withFileTypes: true })) {
      if (entry.isDirectory() && SKIP_DIRS.has(entry.name)) {
        continue;
      }

      visit(join(current, entry.name));
    }
  };

  visit(path);

  return newest;
}

// A candidate is usable only if it actually accepts the flag the renders pass. Asking --help is
// cheaper and safer than parsing a version string, and it answers the real question directly:
// this repo's own docs render with flags no released build has yet.
function accepts(candidate) {
  try {
    const help = execFileSync(candidate, ['config', 'export', 'image', '--help'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    return help.includes(REQUIRED_FLAG);
  } catch {
    return false;
  }
}

// OMP_BIN is an explicit choice and always wins; otherwise the renders use this repo's own build,
// rebuilt whenever the Go sources are newer than it.
//
// Deliberately never falls back to an oh-my-posh on PATH. That is by definition some other build,
// and nothing ties it to the tree being rendered - it silently produced a homepage from a binary
// predating two renderer fixes. The same reasoning covers a src/bin build older than the sources:
// it exists, it accepts every flag, and it still renders yesterday's output. A no-op `go build`
// costs ~0.7s against Go's build cache, so checking is cheaper than being wrong.
function resolveCLI() {
  if (process.env.OMP_BIN) {
    if (!accepts(process.env.OMP_BIN)) {
      throw new Error(`OMP_BIN is set to ${process.env.OMP_BIN}, which does not accept ${REQUIRED_FLAG}.`);
    }

    return process.env.OMP_BIN;
  }

  const built = existsSync(BIN) ? statSync(BIN).mtimeMs : 0;

  if (built && newestMtime(SRC_DIR) <= built) {
    return BIN;
  }

  console.log(built ? 'Go sources are newer than src/bin, rebuilding...' : 'Building oh-my-posh from src...');
  // Same strip flags the release build uses (src/.goreleaser.yml). This binary only ever renders
  // SVGs for the build, so its symbol table and debug info buy nothing.
  execFileSync('go', ['build', '-ldflags=-s -w', '-o', BIN, '.'], { cwd: SRC_DIR, stdio: 'inherit' });

  return BIN;
}

// cmd.exe splits on whitespace, so anything containing a space or quote needs quoting; doubling
// embedded quotes is cmd's own escape convention.
function quoteForCmd(arg) {
  return /[\s"]/.test(arg) ? `"${arg.replace(/"/g, '""')}"` : arg;
}

function run(script, cli) {
  console.log(`Generating ${script}...`);

  const isWindows = process.platform === 'win32';
  // Node deprecates passing an args array alongside shell: true (DEP0190) - it can't escape them
  // for an unknown shell, so on Windows the full, already-quoted command is built as one string
  // instead and given no separate args.
  const [command, args] = isWindows ? [['npm', 'run', script].map(quoteForCmd).join(' '), []] : ['npm', ['run', script]];

  const result = spawnSync(command, args, {
    cwd: WEBSITE_DIR,
    stdio: 'inherit',
    shell: isWindows,
    env: cli ? { ...process.env, OMP_BIN: cli } : process.env,
  });

  if (result.status !== 0) {
    throw new Error(`npm run ${script} failed`);
  }
}

const stale = ARTIFACTS.filter((artifact) => {
  if (artifact.output.some((output) => !existsSync(output))) {
    return true;
  }

  // The oldest output is what decides: a script writing several files is only as up to date as
  // the last one of them to have been written.
  const built = Math.min(...artifact.output.map((output) => statSync(output).mtimeMs));

  return artifact.inputs.some((input) => newestMtime(input) > built);
});

if (!stale.length) {
  process.exit(0);
}

// Failures here are things the developer has to act on - no Go toolchain, an OMP_BIN pointing at
// the wrong binary, a render that died. A stack trace through this script says nothing useful
// about any of them, so print the message and the manual commands instead.
try {
  const cli = stale.some((artifact) => artifact.needsCLI) ? resolveCLI() : null;

  for (const artifact of stale) {
    run(artifact.script, artifact.needsCLI ? cli : null);
  }
} catch (err) {
  console.error(`\nCould not generate the website's build artifacts: ${err.message}\n`);
  console.error('To do it by hand, from the repository root:');
  console.error('  cd src && go build -o ./bin/oh-my-posh . && cd ..');
  console.error('  cd website');

  for (const artifact of stale) {
    console.error(`  npm run ${artifact.script}`);
  }

  process.exit(1);
}
