import React, { useCallback, useEffect, useRef, useState } from 'react';
import ConfigEditor from '../ConfigEditor';
import WasmMessage from '../ConfigEditor/WasmMessage';
import { useWasmRenderer } from '../ConfigEditor/useWasmRenderer';
import { buildRenderOptions, RENDER_DATA_JSON } from '../ConfigEditor/renderDefaults';
import { convertConfig, parseConfig, stringifyConfig } from '../ConfigEditor/serialize';
import {
  APPEND_KEY,
  LOAD_KEY,
  SESSION_KEY,
  trySessionStorageGet,
  trySessionStorageRemove,
  trySessionStorageSet,
} from '../ConfigEditor/studioHandoff';
import { CONFIG_FORMAT, CONFIG_FORMATS, STARTERS } from './config';
import styles from './styles.module.css';

// 120 columns matches the gallery (export_themes.mjs / plugins/themes); see renderDefaults.js
// for the font metrics and sample data this shares with every other live/build-time preview.
const RENDER_OPTIONS = buildRenderOptions(120);

// Rendering itself is local and fast (no network round trip), but re-running it on every
// keystroke would still mean parsing and re-executing every segment's templates on every single
// character typed. 200ms is long enough to coalesce a burst of keystrokes, short enough that the
// preview still feels live.
const DEBOUNCE_MS = 200;

function Studio() {
  const [format, setFormat] = useState(CONFIG_FORMAT);
  const [configText, setConfigText] = useState(STARTERS[CONFIG_FORMAT]);
  // Set once, on mount, if a hand-off from a segment doc's "Add to Studio" couldn't be applied
  // (see the mount effect below) - the append is skipped and this explains why, rather than the
  // segment silently vanishing.
  const [appendNotice, setAppendNotice] = useState(null);
  // Set when a format switch was refused because the config in the editor does not parse; cleared
  // by the next switch that succeeds. See handleFormatChange.
  const [formatNotice, setFormatNotice] = useState(null);

  const { svg, error, wasmStatus, wasmProgress, wasmErrorMessage, render, ensureLoaded } =
    useWasmRenderer({ eager: true });

  const debounceRef = useRef(null);
  // runRender is a useCallback with no deps, so reading `format` from the closure would pin
  // it to whatever was selected on first render. The ref is what the render call reads.
  const formatRef = useRef(CONFIG_FORMAT);

  const runRender = useCallback(
    (text) => {
      render(text, formatRef.current, RENDER_DATA_JSON, RENDER_OPTIONS);
    },
    [render],
  );

  useEffect(() => {
    ensureLoaded();
  }, [ensureLoaded]);

  // Resumes the reader's own session (see studioHandoff.js's SESSION_KEY doc comment), applies a
  // pending theme load if "Open in Studio" queued one (LOAD_KEY, replacing rather than resuming),
  // and otherwise, if a segment doc queued one, appends its segment on top - so "Add to Studio"
  // really does add to whatever is already here rather than to a pristine starter every time.
  // Runs exactly once, on mount: format/configText are intentionally read as their *initial*
  // values (the plain STARTER), not the current state, since this effect IS what decides what
  // "current" becomes.
  useEffect(() => {
    let baseFormat = CONFIG_FORMAT;
    let baseText = STARTERS[CONFIG_FORMAT];

    // LOAD_KEY ("Open in Studio" on a theme) takes priority over everything else: it means
    // "start from this theme", not "add to whatever was already open". A hit here skips the
    // SESSION_KEY resume below entirely (the whole point is to replace it) and the stale
    // APPEND_KEY removal further down guards against an unrelated queued segment from an
    // earlier, unfinished visit suddenly attaching itself to a freshly loaded theme.
    const loadRaw = trySessionStorageGet(LOAD_KEY);
    let loaded = false;

    if (loadRaw) {
      // One-shot regardless of outcome, same reasoning as APPEND_KEY below.
      trySessionStorageRemove(LOAD_KEY);

      try {
        const payload = JSON.parse(loadRaw);

        if (
          payload &&
          typeof payload.text === 'string' &&
          CONFIG_FORMATS.includes(payload.format)
        ) {
          baseFormat = payload.format;
          baseText = payload.text;
          loaded = true;
        }
      } catch {
        // Corrupt payload - fall through to the normal session-resume/starter path below.
      }
    }

    if (!loaded) {
      const savedRaw = trySessionStorageGet(SESSION_KEY);

      if (savedRaw) {
        try {
          const saved = JSON.parse(savedRaw);

          if (saved && typeof saved.text === 'string' && CONFIG_FORMATS.includes(saved.format)) {
            baseFormat = saved.format;
            baseText = saved.text;
          }
        } catch {
          // Corrupt or foreign sessionStorage value - fall back to the pristine starter.
        }
      }
    }

    const pendingRaw = trySessionStorageGet(APPEND_KEY);

    if (loaded) {
      // A theme load supersedes any queued segment append too - discard it rather than silently
      // bolting an unrelated segment onto the theme that was just opened.
      if (pendingRaw) {
        trySessionStorageRemove(APPEND_KEY);
      }
    } else if (pendingRaw) {
      // One-shot: consumed here regardless of what happens next, so a later, unrelated studio
      // visit never re-applies it.
      trySessionStorageRemove(APPEND_KEY);

      try {
        const payload = JSON.parse(pendingRaw);
        const segment = payload && payload.segment;

        if (segment && typeof segment === 'object' && !Array.isArray(segment)) {
          const parsedBase = parseConfig(baseFormat, baseText);
          const blocks = Array.isArray(parsedBase.blocks) ? parsedBase.blocks : [];
          let promptBlock = blocks.find((b) => b && b.type === 'prompt' && Array.isArray(b.segments));

          if (!promptBlock) {
            promptBlock = blocks.find((b) => b && Array.isArray(b.segments));
          }

          if (!promptBlock) {
            promptBlock = { type: 'prompt', alignment: 'left', segments: [] };
            blocks.unshift(promptBlock);
          }

          promptBlock.segments = [...promptBlock.segments, segment];
          parsedBase.blocks = blocks;

          baseText = stringifyConfig(baseFormat, parsedBase);
        }
      } catch (err) {
        // The spec here is explicit: a config that fails to parse must not be silently
        // discarded. Leave baseText exactly as it was (the restored session, or the pristine
        // starter) and surface why the segment wasn't added.
        setAppendNotice(
          `Could not add the segment from the previous page: ${err.message || err}. ` +
            'Your current config was left unchanged.',
        );
      }
    }

    if (baseFormat !== CONFIG_FORMAT) {
      setFormat(baseFormat);
      formatRef.current = baseFormat;
    }

    if (baseText !== STARTERS[CONFIG_FORMAT]) {
      setConfigText(baseText);
    }
    // Deliberately runs once: this is a mount-time migration from sessionStorage into React
    // state, not a reaction to any prop/state change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Saves the reader's own session on every change (see studioHandoff.js's SESSION_KEY doc
  // comment) so navigating away - to read another segment doc, say - and back finds this studio
  // exactly as it was left, which is what lets a second "Add to Studio" hand-off accumulate on
  // top of the first instead of landing on a fresh pristine starter.
  useEffect(() => {
    trySessionStorageSet(SESSION_KEY, JSON.stringify({ format, text: configText }));
  }, [format, configText]);

  // Render the starter config as soon as the module is ready, so a prompt appears on first
  // paint without the user having to touch the editor.
  useEffect(() => {
    if (wasmStatus === 'ready') {
      runRender(configText);
    }
    // configText is intentionally not a dependency: this effect's only job is the one render
    // that happens the moment the module becomes ready, not every keystroke after - the
    // editor's onChange handles that, debounced.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wasmStatus, runRender]);

  useEffect(
    () => () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    },
    [],
  );

  // Switching format rewrites the config into the format that was picked, keeping whatever the
  // reader has typed (see convertConfig). A config that does not parse cannot be rewritten, so
  // the switch is refused rather than left half-applied - a JSON document being read as TOML was
  // the whole bug.
  const handleFormatChange = useCallback(
    (next) => {
      const text = convertConfig(formatRef.current, next, configText);

      if (text === null) {
        setFormatNotice(
          `Could not rewrite this config as ${next}. Fix the errors below and try again.`,
        );
        return;
      }

      setFormatNotice(null);
      setFormat(next);
      formatRef.current = next;
      setConfigText(text);

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      runRender(text);
    },
    [configText, runRender],
  );

  // react-simple-code-editor's onValueChange hands back the raw string (there is no
  // change event to read a value off of), otherwise this is the exact same debounce as before.
  const handleChange = useCallback(
    (value) => {
      setConfigText(value);

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      debounceRef.current = setTimeout(() => runRender(value), DEBOUNCE_MS);
    },
    [runRender],
  );

  const previewHint = wasmStatus === 'ready' && !svg && !error ? 'Nothing to preview yet.' : null;

  return (
    <div className={styles.studio}>
      {/* Preview first: it is what the page is for, and a prompt is short enough
          that putting it above the editor keeps both in view while typing. */}
      <div className={styles.pane}>
        <div className={styles.preview}>
          <WasmMessage
            status={wasmStatus}
            progress={wasmProgress}
            errorMessage={wasmErrorMessage}
            className={styles.wasmMessage}
          />

          {error && <p className={styles.error}>{error}</p>}

          {svg && (
            <span className={styles.svgWrapper} dangerouslySetInnerHTML={{ __html: svg }} />
          )}

          {previewHint && <p className={styles.hint}>{previewHint}</p>}
        </div>
      </div>

      {appendNotice && <p className={styles.notice}>{appendNotice}</p>}
      {formatNotice && <p className={styles.notice}>{formatNotice}</p>}

      <ConfigEditor
        label="Config"
        srLabel="Theme config"
        format={format}
        formats={CONFIG_FORMATS}
        onFormatChange={handleFormatChange}
        value={configText}
        onChange={handleChange}
      />
    </div>
  );
}

export default Studio;
