import React, { useEffect, useId, useRef } from 'react';
import { useColorMode } from '@docusaurus/theme-common';
import classnames from 'classnames';
import { EditorState, Compartment } from '@codemirror/state';
import { EditorView, keymap, lineNumbers } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { bracketMatching, indentOnInput } from '@codemirror/language';
import {
  acceptCompletion,
  autocompletion,
  closeBrackets,
  closeBracketsKeymap,
  completionKeymap,
  completionStatus,
} from '@codemirror/autocomplete';
import { lintKeymap } from '@codemirror/lint';
import { CONFIG_FORMATS } from '../Studio/config';
import { getLanguageExtensions } from './editorExtensions';
import { getThemeExtensions } from './editorTheme';
import { externalErrorExtension, setExternalError } from './externalError';
import styles from './styles.module.css';

// Keyboard behaviour for the format switch, mirroring @docusaurus/theme-classic's own Tabs
// (node_modules/@docusaurus/theme-classic/lib/theme/Tabs/index.js's handleKeydown) closely
// enough that a reader who knows the docs' tabs already knows how to drive this one: Enter/Space
// activates the focused tab, the arrow keys move focus between tabs (wrapping at the ends).
function FormatTabs({ formats, format, onFormatChange }) {
  const itemRefs = useRef([]);
  itemRefs.current = [];

  const focusItem = (index) => {
    const wrapped = (index + formats.length) % formats.length;
    itemRefs.current[wrapped]?.focus();
  };

  const handleKeyDown = (event, index) => {
    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        onFormatChange(formats[index]);
        break;
      case 'ArrowRight':
        event.preventDefault();
        focusItem(index + 1);
        break;
      case 'ArrowLeft':
        event.preventDefault();
        focusItem(index - 1);
        break;
      default:
        break;
    }
  };

  return (
    <ul
      role="tablist"
      aria-orientation="horizontal"
      aria-label="Config format"
      className={classnames('tabs', styles.formatTabs)}
    >
      {formats.map((id, index) => (
        <li
          key={id}
          role="tab"
          ref={(el) => {
            itemRefs.current[index] = el;
          }}
          tabIndex={id === format ? 0 : -1}
          aria-selected={id === format}
          className={classnames('tabs__item', { 'tabs__item--active': id === format })}
          onClick={() => onFormatChange(id)}
          onKeyDown={(event) => handleKeyDown(event, index)}
        >
          {id}
        </li>
      ))}
    </ul>
  );
}

// Extensions every format shares, regardless of which language Compartment content is active:
// undo/redo, auto-indent on newline, bracket matching/closing, and the completion UI itself.
// autocompletion() belongs here rather than inside editorExtensions.js's per-format list -
// schemaCompletion.js's schemaCompletionSource only ever REGISTERS itself via the language's own
// `data.of({ autocomplete })` facet (see editorExtensions.js), it never installs the popup/keymap
// machinery that actually reads it - toml simply never populates that facet, so including this
// unconditionally costs it nothing.
function buildBaseExtensions() {
  return [
    lineNumbers(),
    history(),
    indentOnInput(),
    bracketMatching(),
    closeBrackets(),
    autocompletion(),
    keymap.of([
      // completionKeymap only binds Enter for acceptance; Tab is the other accept key every
      // editor (and this one, before the CodeMirror migration) teaches readers to reach for.
      // acceptCompletion returns false when no completion is active, letting Tab fall through
      // to its default behaviour then.
      { key: 'Tab', run: acceptCompletion },
      ...closeBracketsKeymap,
      ...defaultKeymap,
      ...historyKeymap,
      ...completionKeymap,
      ...lintKeymap,
    ]),
  ];
}

// The reusable half of the studio: a CodeMirror 6 surface with schema-aware completion/linting
// (json, yaml), syntax highlighting (all three formats), and a format switch, extracted so both
// the studio (Studio/index.js) and each segment doc's own editor (Config.js) get the exact same
// editing surface. Rendering (debounce, wasm load, starters, error handling) stays with each
// caller - see useWasmRenderer.js for the piece of that which *is* shared.
//
// A fully controlled component: `value`/`onChange` and `format`/`onFormatChange` are owned by
// the caller, same as before this was migrated off react-simple-code-editor.
function ConfigEditor({
  label = 'Config',
  srLabel,
  format,
  formats = CONFIG_FORMATS,
  onFormatChange,
  value,
  onChange,
  actions = null,
  // 'config' (default) completes against the full theme schema - the studio's own use case.
  // 'segment' completes against #/definitions/segment instead, for the bare-segment sample
  // editor every segment doc page renders (see Config.js) - without this, completion would
  // resolve every top-level property against the wrong schema (root config fields instead of
  // segment fields). See editorExtensions.js's getScopedSchema.
  schemaScope = 'config',
  // { line, column, endColumn } (all 1-based) of a syntax error to draw a squiggly underline
  // under, or null/undefined for none - see errorPosition.js's getSyntaxErrorLocation, which
  // every caller derives this from. Deliberately just one line: a squiggle spanning several
  // lines (an unterminated string, say) would need a far more elaborate overlay for very
  // little benefit over pointing at where the problem starts.
  errorLocation = null,
  // The parse error's own message, shown in a tooltip when hovering the squiggle drawn at
  // errorLocation - null/undefined whenever errorLocation is, so both are always kept in sync
  // by the caller (see Studio/index.js and Config.js's shared `displayError` gate).
  errorMessage = null,
  // Notified with the completion popup's own open/closed state, so a caller can suppress its
  // error banner while it's open - the config is almost always momentarily invalid mid-completion
  // (an open string, a freshly chained empty object), and there is no point flagging that to
  // a reader who is still actively picking a suggestion. 'pending' (a completion source is still
  // being queried) intentionally does NOT count as open - only 'active' does - matching how the
  // old hand-rolled popup only ever reported itself open once it actually had items to show.
  onCompletionOpenChange,
}) {
  const { colorMode } = useColorMode();

  // CodeMirror's own editable surface has no plain <textarea> to attach an aria-label to (see
  // EditorView.contentAttributes below), so a real <label htmlFor> stands in - useId gives each
  // instance its own id, so a page that ever ends up with more than one ConfigEditor (there
  // isn't one today) still gets a valid association per instance instead of a duplicate-id
  // collision.
  const textareaId = useId();
  const hostRef = useRef(null);
  const viewRef = useRef(null);

  // Compartments let format/schemaScope and colorMode changes reconfigure just their own slice
  // of the extension list (see the effects below) without tearing down and recreating the whole
  // EditorView - which would otherwise drop undo history and cursor position on every format
  // switch or dark/light toggle. Lazily created once per instance (not `useRef(new
  // Compartment())`, which would construct a fresh, immediately-discarded Compartment on every
  // render just to seed the ref).
  const languageCompartmentRef = useRef(null);
  if (!languageCompartmentRef.current) {
    languageCompartmentRef.current = new Compartment();
  }
  const themeCompartmentRef = useRef(null);
  if (!themeCompartmentRef.current) {
    themeCompartmentRef.current = new Compartment();
  }

  // onChange/onCompletionOpenChange are read from refs inside the update listener rather than
  // closed over directly, so a caller passing a fresh function identity on every render (neither
  // Studio/index.js nor Config.js memoizes these two) never forces the mount effect below to
  // tear down and recreate the view.
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const onCompletionOpenChangeRef = useRef(onCompletionOpenChange);
  onCompletionOpenChangeRef.current = onCompletionOpenChange;
  // completionStatus(state) is recomputed on every doc/selection change; this only tracks
  // whether the OPEN-ness derived from it actually flipped, so onCompletionOpenChange fires on
  // real transitions instead of once per keystroke while a completion stays open.
  const completionOpenRef = useRef(false);

  // Creates the EditorView exactly once per mount. format/schemaScope/colorMode/value/label are
  // deliberately read here only as INITIAL values - each has its own reconfigure effect below,
  // because recreating the whole view on every prop change would drop undo history and cursor
  // position, the same reason the Compartments above exist at all.
  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return undefined;
    }

    const updateListener = EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        onChangeRef.current?.(update.state.doc.toString());
      }

      const isOpen = completionStatus(update.state) === 'active';
      if (isOpen !== completionOpenRef.current) {
        completionOpenRef.current = isOpen;
        onCompletionOpenChangeRef.current?.(isOpen);
      }
    });

    const view = new EditorView({
      parent: host,
      state: EditorState.create({
        doc: value,
        extensions: [
          ...buildBaseExtensions(),
          languageCompartmentRef.current.of(getLanguageExtensions(format, schemaScope)),
          themeCompartmentRef.current.of(getThemeExtensions(colorMode)),
          externalErrorExtension,
          // No `white-space: pre` class trick needed here (the old react-simple-code-editor
          // setup's .noWrap) - simply never adding EditorView.lineWrapping gets the same
          // one-logical-line-per-visual-row behaviour, with long lines scrolling horizontally
          // (see styles.module.css's .cm-scroller rule) instead of wrapping.
          EditorView.contentAttributes.of({ 'aria-label': srLabel ?? label, id: textareaId }),
          updateListener,
        ],
      }),
    });

    viewRef.current = view;

    return () => {
      view.destroy();
      viewRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Controlled `value`: whenever it no longer matches the document CodeMirror is actually
  // showing (an external change - a format conversion, a "Load"/"Add to Studio" hand-off, or a
  // completion applied by the caller's own onChange round-trip), replace the whole document.
  // Comparing against view.state.doc.toString() (not tracking "did WE just emit this" some other
  // way) is what keeps a normal keystroke a no-op here: onChange already updated the caller's
  // state to match what CodeMirror already has, so this effect's own `value !== ...` check is
  // false and nothing is dispatched back at the view mid-edit.
  useEffect(() => {
    const view = viewRef.current;
    if (!view || value === view.state.doc.toString()) {
      return;
    }

    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: value },
    });
  }, [value]);

  // format/schemaScope together decide which language+schema extensions are active (see
  // editorExtensions.js) - swapping either reconfigures the language Compartment rather than
  // recreating the view.
  useEffect(() => {
    const view = viewRef.current;
    if (!view) {
      return;
    }

    view.dispatch({
      effects: languageCompartmentRef.current.reconfigure(getLanguageExtensions(format, schemaScope)),
    });
  }, [format, schemaScope]);

  // Follows the site's own dark/light toggle (@docusaurus/theme-common), same as the fenced code
  // blocks in the docs (@theme/CodeBlock) - see editorTheme.js for the palette pair this reuses.
  useEffect(() => {
    const view = viewRef.current;
    if (!view) {
      return;
    }

    view.dispatch({
      effects: themeCompartmentRef.current.reconfigure(getThemeExtensions(colorMode)),
    });
  }, [colorMode]);

  // Draws/clears the caller-supplied squiggle - see externalError.js for why this stays a
  // separate decoration rather than feeding into @codemirror/lint's own diagnostic state.
  useEffect(() => {
    const view = viewRef.current;
    if (!view) {
      return;
    }

    setExternalError(view, errorLocation && errorMessage ? { ...errorLocation, message: errorMessage } : null);
  }, [errorLocation, errorMessage]);

  return (
    <div className={styles.pane}>
      {/* No visible pane label: the editor's own window chrome and its format tabs already say
          what this is, and "CONFIG" in front of them only pushed the tabs off the left edge.
          `label` survives as the accessible name for the editor below. */}
      <div className={styles.paneHeader}>
        <FormatTabs formats={formats} format={format} onFormatChange={onFormatChange} />
        {actions && <div className={styles.actions}>{actions}</div>}
      </div>
      <label htmlFor={textareaId} className={styles.srOnly}>
        {srLabel ?? label}
      </label>
      <div className={styles.editorWindow}>
        <div className={styles.editorChrome} aria-hidden="true">
          <span className={styles.chromeControl}>−</span>
          <span className={styles.chromeControl}>□</span>
          <span className={styles.chromeControl}>×</span>
        </div>
        <div className={styles.editorBody}>
          <div ref={hostRef} className={styles.editorHost} />
        </div>
      </div>
    </div>
  );
}

export default ConfigEditor;
