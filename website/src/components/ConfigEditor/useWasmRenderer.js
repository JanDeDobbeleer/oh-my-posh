import { useCallback, useEffect, useRef, useState } from 'react';
import useIsBrowser from '@docusaurus/useIsBrowser';
import useBaseUrl from '@docusaurus/useBaseUrl';
import { loadWasmModule } from './wasmLoader';

// Owns the wasm module's lifecycle (load once, memoized - see wasmLoader.js) and the render call
// itself, shared by the studio and every segment doc's editor (see Studio/index.js and Config.js).
//
// `eager` is the one behavioural knob the two consumers disagree on. The studio loads the module
// the moment it mounts, same as before this was extracted - a reader landed on /docs/studio to
// use the renderer, so there is nothing to save by waiting. A segment doc is different: most
// visitors read it and leave without ever touching the editor, and the whole reason its "Sample
// Configuration" is pre-rendered at build time (see scripts/render-segment-previews.mjs /
// Config.js) is so those visitors never pay for a ~29 MB download. `eager: false` starts the hook
// in an 'idle' status that never transitions on its own; the consumer calls the returned
// `ensureLoaded()` itself, once, on the reader's first edit.
export function useWasmRenderer({ eager = true } = {}) {
  const isBrowser = useIsBrowser();
  const wasmExecUrl = useBaseUrl('/wasm_exec.js');
  const wasmUrl = useBaseUrl('/omp.wasm');

  const [svg, setSvg] = useState(null);
  const [error, setError] = useState(null);
  const [wasmStatus, setWasmStatus] = useState(eager ? 'loading' : 'idle');
  const [wasmProgress, setWasmProgress] = useState(0);
  const [wasmErrorMessage, setWasmErrorMessage] = useState('');

  // A ref, not state: the render function itself never needs to trigger a re-render when it
  // changes - only wasmStatus flipping to 'ready' does, and that already happens separately.
  const renderRef = useRef(null);
  const startedRef = useRef(false);
  const unmountedRef = useRef(false);

  useEffect(
    () => () => {
      unmountedRef.current = true;
    },
    [],
  );

  // Idempotent and safe to call from either consumer at any time: the first call - whether from
  // the studio's own mount effect below or a segment editor's first keystroke - wins, and every
  // later call is a no-op. loadWasmModule's own module-scope promise (wasmLoader.js) means even
  // two *different* ConfigEditor-backed pages calling this independently still share one fetch.
  const ensureLoaded = useCallback(() => {
    if (!isBrowser || startedRef.current) {
      return;
    }

    startedRef.current = true;
    setWasmStatus('loading');

    loadWasmModule(wasmExecUrl, wasmUrl, (fraction) => {
      if (!unmountedRef.current) {
        setWasmProgress(fraction);
      }
    })
      .then((renderFn) => {
        if (unmountedRef.current) {
          return;
        }

        renderRef.current = renderFn;
        setWasmStatus('ready');
      })
      .catch((err) => {
        if (unmountedRef.current) {
          return;
        }

        setWasmErrorMessage(err.message || String(err));
        setWasmStatus('error');
      });
  }, [isBrowser, wasmExecUrl, wasmUrl]);

  useEffect(() => {
    if (eager) {
      ensureLoaded();
    }
    // Only ever fires the one eager load; ensureLoaded's own startedRef guard is what actually
    // prevents a second load, so re-running this on every ensureLoaded identity change is safe
    // and unnecessary either way.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [eager]);

  const render = useCallback((text, format, dataJson, options) => {
    if (!renderRef.current) {
      return;
    }

    let result;

    try {
      result = renderRef.current(text, format, dataJson, options);
    } catch (err) {
      // renderJS (src/wasm/main.go) is documented to always return {error} rather than throw,
      // but this is still untrusted-from-JS's-perspective wasm glue - fail closed rather than
      // let an unexpected exception escape.
      setError(err.message || String(err));
      return;
    }

    if (result && typeof result.error === 'string') {
      // Deliberately not touching `svg` here: an invalid half-typed config should leave the
      // last good preview on screen instead of blanking it.
      setError(result.error);
      return;
    }

    if (result && typeof result.svg === 'string') {
      setSvg(result.svg);
      setError(null);
      return;
    }

    setError('The renderer returned an unexpected response.');
  }, []);

  return { svg, error, wasmStatus, wasmProgress, wasmErrorMessage, render, ensureLoaded };
}
