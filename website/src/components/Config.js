import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import CodeBlock from '@theme/CodeBlock';
import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";
import Link from '@docusaurus/Link';
import classnames from 'classnames';
import YAML from 'yaml';
import TOML from 'smol-toml';
import { useDoc } from '@docusaurus/plugin-content-docs/client';
import ConfigEditor from './ConfigEditor';
import WasmMessage from './ConfigEditor/WasmMessage';
import { useWasmRenderer } from './ConfigEditor/useWasmRenderer';
import { buildRenderOptions, RENDER_DATA_JSON } from './ConfigEditor/renderDefaults';
import { convertConfig, parseConfig, stringifyConfig } from './ConfigEditor/serialize';
import { APPEND_KEY, trySessionStorageSet } from './ConfigEditor/studioHandoff';
import styles from './Config.module.css';

// The build-time render of every segment doc's own "Sample Configuration" (see
// scripts/render-segment-previews.mjs), keyed by doc id. A plain static import - like
// ThemeGallery's own generated/themes.json - rather than a usePluginData plugin: this file is
// dozens of inlined SVGs, the same order of magnitude plugins/themes/index.js measured at ~34KB
// gzip and rejected usePluginData for, because setGlobalData bundles a plugin's entire payload
// into Docusaurus's shared main.js runtime chunk, fetched on every page of the whole site
// (including the blog and every non-segment doc). Config.js is used on 135 pages, so a static
// import here is a bigger bet than ThemeGallery's single page - but it still only costs the
// pages that render a <Config/> at all, via webpack's own code-splitting, rather than every
// visitor on every page forever.
import previews from '../../generated/segment-previews.json';

const SEGMENT_FORMATS = ['json', 'yaml', 'toml'];
const SEGMENT_DEFAULT_FORMAT = 'json';

// Matches scripts/render-segment-previews.mjs's PREVIEW_COLUMNS, so a segment's live preview
// lines up with the build-time static SVG (generated/segment-previews.json) it sits in for.
const SEGMENT_PREVIEW_COLUMNS = 64;
const SEGMENT_RENDER_OPTIONS = buildRenderOptions(SEGMENT_PREVIEW_COLUMNS);

// Same rationale as the studio's own debounce (Studio/index.js) - long enough to coalesce a
// burst of keystrokes, short enough that the preview still feels live.
const DEBOUNCE_MS = 200;

// Mirrors scripts/render-segment-previews.mjs's buildPromptConfig: wraps a single segment as the
// lone segment of a minimal version-4, one-block config, which is the shape the wasm renderer
// (and the CLI, at build time) actually accepts - see src/wasm/main.go / config.ParseBytes. The
// editor itself only ever shows the bare segment (see segmentStarters above, and the Copy action
// below) since that is what a reader pastes into their own segments list; wrapping only happens
// for this live-preview render call.
function buildPromptConfig(segment) {
  return {
    version: 4,
    final_space: true,
    blocks: [{ type: 'prompt', alignment: 'left', segments: [segment] }],
  };
}

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef(null);

  useEffect(
    () => () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    },
    [],
  );

  const handleClick = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);

      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard API unavailable or permission denied - nothing sensible to surface beyond
      // leaving the button as it was; the text is still selectable by hand either way.
    }
  }, [text]);

  return (
    <button type="button" className={styles.action} onClick={handleClick}>
      {copied ? 'Copied' : 'Copy'}
    </button>
  );
}

// Hands the current segment off to the studio via sessionStorage (see ConfigEditor/
// studioHandoff.js) and lets the <Link> perform its normal navigation to /docs/studio - the
// studio reads and clears the hand-off on its own mount. Disabled (inert, but still focusable)
// while the current text does not parse: there is no valid segment object to hand off, and per
// spec a failed hand-off must not silently discard anything, so the safer choice is to not
// attempt one at all.
function AddToStudioButton({ format, configText, disabled }) {
  const handleClick = useCallback(
    (event) => {
      if (disabled) {
        event.preventDefault();
        return;
      }

      try {
        const segment = parseConfig(format, configText);
        trySessionStorageSet(APPEND_KEY, JSON.stringify({ segment }));
      } catch {
        // Guarded by `disabled` above for the expected case; if this still throws, degrade to
        // "just navigate" (per the sessionStorage-unavailable guidance) rather than trap the
        // reader on this page.
      }
    },
    [format, configText, disabled],
  );

  return (
    <Link
      to="/docs/studio"
      className={classnames(styles.action, { [styles.actionDisabled]: disabled })}
      aria-disabled={disabled}
      onClick={handleClick}
    >
      Add to Studio
    </Link>
  );
}

// The editable half of a segment doc's "Sample Configuration": the build-time static SVG shows
// immediately (see Config below), and only turns into a live one once the reader edits the
// config for the first time - see useWasmRenderer's `eager` doc comment for why that matters.
function EditableConfig({ data, staticSvg }) {
  // Only the default format's text is built up front. The other two used to be pre-rendered as
  // well, to be swapped in whole when the reader picked a different tab; handleFormatChange now
  // rewrites whatever is in the editor instead (see convertConfig), so they were never read.
  const initialText = useMemo(() => stringifyConfig(SEGMENT_DEFAULT_FORMAT, data), [data]);

  const [format, setFormat] = useState(SEGMENT_DEFAULT_FORMAT);
  const [configText, setConfigText] = useState(initialText);
  const [parseError, setParseError] = useState(null);

  const formatRef = useRef(SEGMENT_DEFAULT_FORMAT);
  const debounceRef = useRef(null);

  const { svg, error, wasmStatus, wasmProgress, wasmErrorMessage, render, ensureLoaded } =
    useWasmRenderer({ eager: false });

  const runRender = useCallback(
    (text, fmt) => {
      let segment;

      try {
        segment = parseConfig(fmt, text);
      } catch (err) {
        setParseError(err.message || String(err));
        return;
      }

      setParseError(null);
      const wrappedText = stringifyConfig(fmt, buildPromptConfig(segment));
      render(wrappedText, fmt, RENDER_DATA_JSON, SEGMENT_RENDER_OPTIONS);
    },
    [render],
  );

  // Mirrors the studio's own "render as soon as the module is ready" effect (Studio/index.js):
  // the module only starts loading on the reader's first edit (see handleChange below), so by
  // the time it actually becomes ready there is already something worth rendering.
  useEffect(() => {
    if (wasmStatus === 'ready') {
      runRender(configText, format);
    }
    // configText/format intentionally excluded: this effect's only job is the one render that
    // happens the moment the module becomes ready, not every keystroke after.
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

  // Rewrites the segment into the format that was picked, keeping the reader's edits - see
  // convertConfig. A segment caught mid-edit cannot be rewritten, so the switch is refused and
  // the parse error already on screen says why, rather than the tab moving and the editor then
  // reading JSON as TOML.
  const handleFormatChange = useCallback(
    (next) => {
      const text = convertConfig(formatRef.current, next, configText);

      if (text === null) {
        return;
      }

      setFormat(next);
      formatRef.current = next;
      setConfigText(text);

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      // Switching format is not "editing" - the module must not load just because a reader
      // looked at a different tab (see useWasmRenderer's `eager` doc comment). Only refresh the
      // preview if it already happens to be ready from an earlier edit.
      if (wasmStatus === 'ready') {
        runRender(text, next);
      }
    },
    [configText, wasmStatus, runRender],
  );

  // react-simple-code-editor's onValueChange hands back the raw string - same debounce as the
  // studio's, plus the one-time trigger that starts the wasm download.
  const handleChange = useCallback(
    (value) => {
      setConfigText(value);
      ensureLoaded();

      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      debounceRef.current = setTimeout(() => runRender(value, formatRef.current), DEBOUNCE_MS);
    },
    [runRender, ensureLoaded],
  );

  // Keeps the last good preview on screen instead of blanking it: before the first edit (or
  // while the module is still loading) that is the build-time static SVG; once a live render has
  // actually succeeded at least once, its own svg takes over and is itself kept across a later
  // parse/render error, same as the studio.
  const displaySvg = svg || staticSvg;
  const displayError = parseError || error;

  return (
    <div className={styles.editableConfig}>
      <div className={styles.preview}>
        <WasmMessage
          status={wasmStatus}
          progress={wasmProgress}
          errorMessage={wasmErrorMessage}
          className={styles.wasmMessage}
        />
        {displayError && <p className={styles.error}>{displayError}</p>}
        {displaySvg && (
          <span className={styles.svgWrapper} dangerouslySetInnerHTML={{ __html: displaySvg }} />
        )}
      </div>
      <ConfigEditor
        label="Config"
        srLabel={`${data.type} segment config`}
        format={format}
        formats={SEGMENT_FORMATS}
        onFormatChange={handleFormatChange}
        value={configText}
        onChange={handleChange}
        actions={
          <>
            <CopyButton text={configText} />
            <AddToStudioButton format={format} configText={configText} disabled={!!parseError} />
          </>
        }
      />
    </div>
  );
}

function Config(props) {

  const { data, metastring = { json: "", yaml: "", toml: "" } } = props;
  // frontMatter.id, not metadata.id: every segment doc sets an explicit `id:` frontmatter field
  // (e.g. "kubectl"), but Docusaurus's own metadata.id is the full sourceDirName-qualified id
  // (e.g. "segments/cli/kubectl") regardless of that override. generated/segment-previews.json is
  // keyed on the plain frontmatter id (see scripts/render-segment-previews.mjs), matching what
  // every other build-time consumer of these docs already keys on (plugins/segments/index.js's
  // registry, extract-segment-properties.mjs).
  const { frontMatter } = useDoc();
  const entry = previews[frontMatter.id];
  // Six segment docs render a second, illustrative <Config/> further down the page (e.g.
  // scm/git.mdx's "posh-git" section) with the same doc id but different sample data - the
  // manifest only ever describes the doc's first ("## Sample Configuration") block, so a bare
  // id lookup would wrongly attach that same preview to the unrelated second block too. Comparing
  // the stored sample object against this instance's own `data` prop confirms this really is the
  // <Config/> the preview was rendered from - and, now, which instance gets to become editable:
  // an unmatched second block keeps the plain tabs below rather than turning into its own editor.
  const preview = entry && JSON.stringify(entry.segment) === JSON.stringify(data) ? entry.svg : null;

  if (preview) {
    return <EditableConfig data={data} staticSvg={preview} />;
  }

  const patchTomlData = () => {
    if (data?.properties) {
      const properties = data.properties;
      delete data.properties;

      return {
        ...data,
        blocks: {
          segments: {
            properties: properties
          }
        }
      };
    }

    return data;
  };

  return (
    <>
      <Tabs
        defaultValue="json"
        groupId="sample"
        values={[
          { label: 'json', value: 'json', },
          { label: 'yaml', value: 'yaml', },
          { label: 'toml', value: 'toml', },
        ]
      }>
        <TabItem value="json">
          <CodeBlock language="json" metastring={metastring.json}>
            {JSON.stringify(data, null, 2)}
          </CodeBlock>
        </TabItem>
        <TabItem value="yaml">
          <CodeBlock language="yaml" metastring={metastring.yaml}>
            {YAML.stringify(data)}
          </CodeBlock>
        </TabItem>
        <TabItem value="toml">
          <CodeBlock language="toml" metastring={metastring.toml}>
            {TOML.stringify(patchTomlData())}
          </CodeBlock>
        </TabItem>
      </Tabs>
    </>
  );
}

export default Config;
