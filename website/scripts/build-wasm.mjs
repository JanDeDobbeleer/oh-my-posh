#!/usr/bin/env node
// Builds the studio's Go/wasm module (src/wasm) into website/static/omp.wasm and copies the
// matching Go host shim (wasm_exec.js) beside it. A plain Node script rather than a shell one-
// liner in package.json: GOOS/GOARCH assignment, `cd`, and copying a file out of `go env GOROOT`
// are each spelled differently between POSIX shells and cmd/PowerShell, and this repo's
// developers and CI (ubuntu-latest) run both. Node's child_process/fs give one implementation
// that works the same everywhere `npm run wasm` runs.
//
// Both outputs are gitignored (see website/.gitignore) and regenerated on every build, exactly
// like generated/themes.json - see this script's own npm script ("wasm") and the "Render themes"
// step it sits beside in .github/workflows/docs.yml.
import { execFileSync } from 'node:child_process';
import { existsSync, copyFileSync, mkdirSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEBSITE_DIR = join(__dirname, '..');
const SRC_DIR = join(WEBSITE_DIR, '..', 'src');
const STATIC_DIR = join(WEBSITE_DIR, 'static');
const WASM_OUT = join(STATIC_DIR, 'omp.wasm');
const WASM_EXEC_OUT = join(STATIC_DIR, 'wasm_exec.js');

function formatBytes(bytes) {
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function buildWasm() {
  mkdirSync(STATIC_DIR, { recursive: true });

  console.log('Building src/wasm for js/wasm...');

  execFileSync('go', ['build', '-ldflags=-s -w', '-trimpath', '-o', WASM_OUT, './wasm/'], {
    cwd: SRC_DIR,
    // GOOS/GOARCH here play the same role as the CI step's GOOS=js GOARCH=wasm - src/go.mod's
    // own module is otherwise built for the host platform, which src/wasm/main.go's `//go:build
    // js && wasm` constraint would then exclude entirely, leaving nothing for `go build` to do.
    //
    // -s -w are the release build's own strip flags (see src/.goreleaser.yml): drop the symbol
    // table and DWARF debug info. Every visitor to the studio downloads this file, so the debug
    // information a browser can do nothing with is pure transfer cost. -trimpath additionally
    // keeps the building machine's absolute paths out of it.
    //
    // Deliberately not carrying the release's `timetzdata` tag, which embeds the IANA database
    // (~450KB) so time.LoadLocation works without system tzdata. The only caller is
    // template/date.go, and it already falls back to UTC when the lookup fails, so a studio
    // config naming a timezone renders UTC instead of erroring - not worth 450KB on every visit.
    env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
    stdio: 'inherit',
  });

  const { size } = statSync(WASM_OUT);
  console.log(`Wrote ${WASM_OUT} (${formatBytes(size)})`);
}

function copyWasmExec() {
  const goroot = execFileSync('go', ['env', 'GOROOT']).toString().trim();
  // Go 1.24+ moved the host shim from misc/wasm/ to lib/wasm/ (this repo builds with Go 1.26 -
  // see src/go.mod). Only the new location is supported; there is nothing here to build with an
  // older Go.
  const wasmExecSrc = join(goroot, 'lib', 'wasm', 'wasm_exec.js');

  if (!existsSync(wasmExecSrc)) {
    throw new Error(`wasm_exec.js not found at ${wasmExecSrc} (GOROOT=${goroot})`);
  }

  copyFileSync(wasmExecSrc, WASM_EXEC_OUT);
  console.log(`Copied ${wasmExecSrc} -> ${WASM_EXEC_OUT}`);
}

buildWasm();
copyWasmExec();
