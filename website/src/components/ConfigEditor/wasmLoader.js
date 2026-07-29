// Loads the studio's wasm module: the Go host shim (wasm_exec.js, a plain script that defines
// window.Go) followed by the module itself (omp.wasm), fetched with byte-level progress so the
// loading state in useWasmRenderer.js can say something more useful than a bare spinner while a
// ~29 MB file downloads.
//
// The whole load is memoized behind a single module-scope promise. Without that, remounting a
// consumer (e.g. React 18/19 StrictMode's double-invoke in dev, or navigating away and back)
// would re-fetch and re-instantiate the module, and - worse - call Go's run() a second time,
// leaving a second Go runtime (and its own permanently-parked main goroutine, see
// src/wasm/main.go's doc comment on why main never returns) alive in the same tab. This also
// means the studio and a segment doc's editor, if both mounted, genuinely share one download.
let wasmModulePromise = null;

function loadScript(src) {
  return new Promise((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve();
      return;
    }

    const script = document.createElement('script');
    script.src = src;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`failed to load ${src}`));
    document.head.appendChild(script);
  });
}

// Reads the response body in chunks instead of a single response.arrayBuffer() so onProgress can
// report how much of the file has arrived - the only reason this is more than one line. Falls
// back to a no-progress single await when the runtime has no streaming body (very old browsers)
// or the server didn't send a Content-Length.
async function fetchWithProgress(url, onProgress) {
  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`failed to download ${url}: HTTP ${response.status}`);
  }

  if (!response.body) {
    return response.arrayBuffer();
  }

  const total = Number(response.headers.get('content-length')) || 0;
  const reader = response.body.getReader();
  const chunks = [];
  let received = 0;
  let lastPct = -1;

  for (;;) {
    const { done, value } = await reader.read();

    if (done) {
      break;
    }

    chunks.push(value);
    received += value.length;

    if (total > 0) {
      const pct = Math.floor((received / total) * 100);

      if (pct !== lastPct) {
        lastPct = pct;
        onProgress(received / total);
      }
    }
  }

  const buffer = new Uint8Array(received);
  let offset = 0;

  for (const chunk of chunks) {
    buffer.set(chunk, offset);
    offset += chunk.length;
  }

  return buffer.buffer;
}

// loadWasmModule resolves to the render(configText, format, dataJSON, optionsObject) function
// src/wasm/main.go registers as a JS global. onProgress is called with a 0..1 fraction as
// omp.wasm downloads (best-effort - see fetchWithProgress).
export function loadWasmModule(wasmExecUrl, wasmUrl, onProgress) {
  if (wasmModulePromise) {
    return wasmModulePromise;
  }

  wasmModulePromise = (async () => {
    await loadScript(wasmExecUrl);

    const go = new window.Go();
    const bytes = await fetchWithProgress(wasmUrl, onProgress);
    const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

    // Not awaited: main() (see src/wasm/main.go) blocks forever in `select{}` so this run()'s
    // returned promise never resolves in normal operation. The synchronous part of run() -
    // which includes main() executing up to that select{} - completes before go.run() returns
    // control here, so window.render is already set by the time the next line runs.
    go.run(instance).catch((err) => {
      // eslint-disable-next-line no-console
      console.error('oh-my-posh wasm runtime exited unexpectedly', err);
    });

    return window.render;
  })().catch((err) => {
    // Let a failed load be retried on the next mount instead of caching the rejection forever.
    wasmModulePromise = null;
    throw err;
  });

  return wasmModulePromise;
}
