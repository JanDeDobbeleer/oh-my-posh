import React, { useCallback, useId, useMemo, useRef } from 'react';
import useIsomorphicLayoutEffect from '@docusaurus/useIsomorphicLayoutEffect';
import { useColorMode } from '@docusaurus/theme-common';
import { Highlight, Prism, themes } from 'prism-react-renderer';
import classnames from 'classnames';
import Editor from 'react-simple-code-editor';
import { CONFIG_FORMATS } from '../Studio/config';
import styles from './styles.module.css';

// The docs' own code fences (@theme/CodeBlock) run on prism-react-renderer too, through
// @docusaurus/theme-common's usePrismTheme() - which today resolves to the *same* palenight
// theme in light and dark mode alike, because docusaurus.config.js's themeConfig.prism never
// sets a `theme`/`darkTheme` pair (confirmed by inspecting a rendered docs code block: its
// computed background is #292d3e and its punctuation #c792ea, palenight's colors - not the
// #232136/#3e8fb0 that website/src/css/prism-rose-pine-moon.css would produce, which turns out
// to be dead CSS for these React-rendered blocks, overridden by prism-react-renderer's own
// inline per-token styling). Reusing palenight for dark mode keeps the editor pixel-identical to
// a real doc's code fence; github stands in as the light counterpart so the editor - unlike the
// docs' code blocks, which stay dark either way - actually adapts with the site's light/dark
// switch, which is the more useful behaviour for something people type into for minutes at a
// time rather than glance at.
const DARK_CODE_THEME = themes.palenight;
const LIGHT_CODE_THEME = themes.github;

// Kept equal to the gutter's own vertical padding (--omp-space-2 in styles.module.css) so
// line N's number sits on line N's code. Expressed here rather than in CSS because the
// editor's inner layers need it inline - see the padding prop below.
const EDITOR_PADDING = '0.75rem';

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

// The reusable half of the studio: a highlighted, gutter-numbered, window-chromed config editor
// with a format switch, extracted so both the studio (Studio/index.js) and each segment doc's
// own editor (Config.js) get the exact same editing surface. Rendering (debounce, wasm load,
// starters, error handling) stays with each caller - see useWasmRenderer.js for the piece of
// that which *is* shared.
//
// A fully controlled component: `value`/`onChange` and `format`/`onFormatChange` are owned by
// the caller, same as before this was extracted out of Studio/index.js.
function ConfigEditor({
  label = 'Config',
  srLabel,
  format,
  formats = CONFIG_FORMATS,
  onFormatChange,
  value,
  onChange,
  actions = null,
}) {
  const { colorMode } = useColorMode();
  const codeTheme = colorMode === 'dark' ? DARK_CODE_THEME : LIGHT_CODE_THEME;

  // react-simple-code-editor's inner <textarea> can't take a plain aria-label prop, so a real
  // <label htmlFor> stands in - useId gives each instance its own id, so a page that ever ends
  // up with more than one ConfigEditor (there isn't one today) still gets a valid association
  // per instance instead of a duplicate-id collision.
  const textareaId = useId();
  const gutterRef = useRef(null);

  // One entry per logical line of the config, independent of the (debounced) render - typing a
  // newline should renumber the gutter immediately, not on the caller's own debounce timer. Safe
  // to derive straight from `value` because .noWrap forces one visual row per logical line, so
  // "line N" always means the same thing here as it does in the highlighted code below.
  const lineCount = useMemo(() => value.split('\n').length, [value]);
  const lineNumbers = useMemo(
    () => Array.from({ length: lineCount }, (_, i) => i + 1),
    [lineCount],
  );

  // Keeps the gutter scrolled in step with .editor, which react-simple-code-editor renders as a
  // plain scrollable div (see styles.module.css's .editor). The gutter is a sibling, not a child,
  // of that div (see the .gutter comment in styles.module.css), so its scroll position has to be
  // mirrored by hand - the editor only scrolls internally when a reader has dragged its resize
  // handle shorter than the config, but when that happens the numbers must move with the code.
  //
  // The gutter's *height* is not set here. It used to be, back when .editor had a fixed
  // min-height; a ResizeObserver wrote container.clientHeight onto the gutter. That write resized
  // .editorBody, which resized .editor, which re-notified the observer - and Chrome's loop guard
  // dropped the follow-up, so a gutter pinned to the old height clipped its last lines the moment
  // the editor grew. Now that .editor sizes itself to its content, .editorBody's own
  // align-items: stretch gives the gutter the row's height, for free and always in step.
  useIsomorphicLayoutEffect(() => {
    const textarea = document.getElementById(textareaId);
    const gutter = gutterRef.current;

    if (!textarea || !gutter) {
      return undefined;
    }

    const container = textarea.parentElement;

    if (!container) {
      return undefined;
    }

    const sync = () => {
      gutter.scrollTop = container.scrollTop;
    };

    sync();
    container.addEventListener('scroll', sync);

    return () => {
      container.removeEventListener('scroll', sync);
    };
  }, [textareaId]);

  // Tokenizes with the exact same prism-react-renderer pipeline @theme/CodeBlock uses for every
  // fenced code block in the docs (see the DARK_CODE_THEME/LIGHT_CODE_THEME comment above), just
  // rendered as plain <div>/<span> lines instead of Docusaurus's own Pre/Code/Line wrappers.
  // The format strings config.ParseBytes takes are also Prism's language ids for all three,
  // so the selected format doubles as the highlight language with no mapping.
  const highlightConfig = useCallback(
    (code) => (
      <Highlight prism={Prism} language={format} code={code} theme={codeTheme}>
        {({ tokens, getLineProps, getTokenProps }) => (
          <>
            {tokens.map((line, i) => (
              <div key={i} {...getLineProps({ line })}>
                {line.map((token, key) => (
                  <span key={key} {...getTokenProps({ token })} />
                ))}
              </div>
            ))}
          </>
        )}
      </Highlight>
    ),
    [format, codeTheme],
  );

  return (
    <div className={styles.pane}>
      {/* No visible pane label: the editor's own window chrome and its format tabs already say
          what this is, and "CONFIG" in front of them only pushed the tabs off the left edge.
          `label` survives as the accessible name for the textarea below. */}
      <div className={styles.paneHeader}>
        <FormatTabs formats={formats} format={format} onFormatChange={onFormatChange} />
        {actions && <div className={styles.actions}>{actions}</div>}
      </div>
      <label htmlFor={textareaId} className={styles.srOnly}>
        {srLabel ?? label}
      </label>
      <div className={styles.editorWindow}>
        <div className={styles.editorChrome} aria-hidden="true">
          <span className={styles.chromeDot} />
          <span className={styles.chromeDot} />
          <span className={styles.chromeDot} />
        </div>
        <div className={styles.editorBody}>
          <div
            ref={gutterRef}
            className={styles.gutter}
            aria-hidden="true"
            style={{ backgroundColor: codeTheme.plain.backgroundColor }}
          >
            {lineNumbers.map((n) => (
              <div key={n} className={styles.gutterLine}>
                {n}
              </div>
            ))}
          </div>
          <Editor
            className={styles.editor}
            preClassName={styles.noWrap}
            textareaClassName={styles.noWrap}
            style={{
              overflow: 'auto',
              backgroundColor: codeTheme.plain.backgroundColor,
              color: codeTheme.plain.color,
            }}
            textareaId={textareaId}
            value={value}
            onValueChange={onChange}
            highlight={highlightConfig}
            // The padding has to come through this prop rather than .editor's CSS.
            // react-simple-code-editor absolutely positions its own <pre> and <textarea>
            // to fill the container, and an absolutely positioned box anchors to the
            // padding box - so CSS padding on .editor moved nothing, leaving the code
            // flush against the gutter's border while the gutter, an ordinary flex
            // sibling with the same padding, sat a line lower. This prop is applied
            // inline to both layers, so the code and the numbers share one origin.
            padding={EDITOR_PADDING}
          />
        </div>
      </div>
    </div>
  );
}

export default ConfigEditor;
