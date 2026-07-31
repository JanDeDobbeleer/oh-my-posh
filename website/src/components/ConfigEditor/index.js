import React, { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import useIsomorphicLayoutEffect from '@docusaurus/useIsomorphicLayoutEffect';
import { useColorMode } from '@docusaurus/theme-common';
import { Highlight, Prism, themes } from 'prism-react-renderer';
import classnames from 'classnames';
import Editor from 'react-simple-code-editor';
import { CONFIG_FORMATS } from '../Studio/config';
import { getCompletions, getCompletionReplacement, getHoverInfo } from './completion';
import styles from './styles.module.css';

// The docs' own code fences (@theme/CodeBlock) run on prism-react-renderer too, through
// @docusaurus/theme-common's usePrismTheme() - which now resolves to this exact pair via
// docusaurus.config.js's themeConfig.prism.theme/darkTheme, so a doc's fenced code and the
// editor stay visually identical in both colour modes.
const DARK_CODE_THEME = themes.palenight;
const LIGHT_CODE_THEME = themes.github;

// Kept equal to the gutter's own vertical padding (--omp-space-2 in styles.module.css) so
// line N's number sits on line N's code. Expressed here rather than in CSS because the
// editor's inner layers need it inline - see the padding prop below.
const EDITOR_PADDING = '0.75rem';

// The editor's font is monospace (.editor in styles.module.css), so every character occupies
// the same width - measuring one digit via a scratch canvas gives the per-column pixel width
// needed to turn a mouse position into a text offset (see handleEditorMouseMove), without
// creating a real DOM node just to measure. Cached per font string since the editor's
// font-size can change across the responsive breakpoint in styles.module.css.
const charWidthCache = new Map();
function getCharWidth(target) {
  const cs = getComputedStyle(target);
  const font = `${cs.fontStyle} ${cs.fontWeight} ${cs.fontSize} ${cs.fontFamily}`;
  const cached = charWidthCache.get(font);
  if (cached) {
    return cached;
  }

  const canvas = getCharWidth.canvas || (getCharWidth.canvas = document.createElement('canvas'));
  const ctx = canvas.getContext('2d');
  ctx.font = font;
  const width = ctx.measureText('0').width;
  charWidthCache.set(font, width);
  return width;
}

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
  // 'config' (default) completes against the full theme schema - the studio's own use case.
  // 'segment' completes against #/definitions/segment instead, for the bare-segment sample
  // editor every segment doc page renders (see Config.js) - without this, completion would
  // resolve every top-level property against the wrong schema (root config fields instead of
  // segment fields).
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
  // Notified with the popup's own open/closed state, so a caller can suppress its error
  // banner while it's open - the config is almost always momentarily invalid mid-completion
  // (an open string, a freshly chained empty object), and there is no point flagging that to
  // a reader who is still actively picking a suggestion.
  onCompletionOpenChange,
}) {
  const { colorMode } = useColorMode();
  const codeTheme = colorMode === 'dark' ? DARK_CODE_THEME : LIGHT_CODE_THEME;

  // react-simple-code-editor's inner <textarea> can't take a plain aria-label prop, so a real
  // <label htmlFor> stands in - useId gives each instance its own id, so a page that ever ends
  // up with more than one ConfigEditor (there isn't one today) still gets a valid association
  // per instance instead of a duplicate-id collision.
  const textareaId = useId();
  const gutterRef = useRef(null);
  const editorBodyRef = useRef(null);
  const activeCompletionRef = useRef(null);
  const [completionState, setCompletionState] = useState({
    open: false,
    items: [],
    selectedIndex: 0,
    top: 0,
    left: 0,
    bottom: 0,
    openUpward: false,
    maxHeight: null,
    replaceStart: 0,
    insideOpenString: false,
  });
  // A separate piece of state from completionState so hovering doesn't fight the
  // keyboard-driven selectedIndex, and so it can be cleared independently (mouse leaves
  // the popup, or the popup itself closes) without re-running completion logic.
  const [tooltip, setTooltip] = useState({ open: false, title: '', text: '', examples: null, top: 0, left: 0 });

  // Keeps the highlighted suggestion visible as arrow keys move past the popup's own
  // max-height/overflow bounds.
  useEffect(() => {
    if (completionState.open) {
      activeCompletionRef.current?.scrollIntoView({ block: 'nearest' });
    }
  }, [completionState.open, completionState.selectedIndex]);

  useEffect(() => {
    onCompletionOpenChange?.(completionState.open);
  }, [completionState.open, onCompletionOpenChange]);

  // Keeps the richer-description tooltip pinned to whichever option is actually highlighted
  // (completionState.selectedIndex - moved by ArrowUp/ArrowDown or by hovering, see the option
  // row's onMouseEnter below), rather than wherever the mouse last happened to be - the previous
  // per-row onMouseEnter/onMouseLeave pair let the tooltip describe a different item than the one
  // the popup was visually highlighting, which was the actual bug being fixed here. Runs after
  // the scrollIntoView effect above, so if the active row just scrolled into view this measures
  // its final on-screen position, not its pre-scroll one.
  useEffect(() => {
    if (!completionState.open) {
      return;
    }

    const item = completionState.items[completionState.selectedIndex];
    const optionElement = activeCompletionRef.current;

    if (!item || !optionElement || (!item.description && !item.examples)) {
      setTooltip((current) => (current.open ? { ...current, open: false } : current));
      return;
    }

    const editorBodyRect = editorBodyRef.current?.getBoundingClientRect();
    if (!editorBodyRect) {
      return;
    }

    const optionRect = optionElement.getBoundingClientRect();
    const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16;
    const tooltipWidth = 18 * rootFontSize;
    const gap = 8;

    // Prefer opening to the right of the popup; flip to the left when there isn't room,
    // same flip-on-overflow idea as the popup's own vertical placement.
    const popupRect = optionElement.parentElement?.getBoundingClientRect();
    const popupRight = popupRect?.right ?? optionRect.right;
    const openLeft = popupRight + gap + tooltipWidth > window.innerWidth;
    const left = openLeft
      ? (popupRect?.left ?? optionRect.left) - editorBodyRect.left - gap - tooltipWidth
      : popupRight - editorBodyRect.left + gap;

    // Keep the tooltip's top edge aligned with the highlighted row, clamped so a row near the
    // bottom of the viewport doesn't push the tooltip off-screen.
    const rawTop = optionRect.top - editorBodyRect.top;
    const maxTop = window.innerHeight - editorBodyRect.top - 8 * rootFontSize;
    const top = Math.min(rawTop, Math.max(0, maxTop));

    setTooltip({
      open: true,
      title: item.detail || item.label,
      text: item.description,
      examples: item.examples,
      top,
      left,
    });
  }, [completionState.open, completionState.items, completionState.selectedIndex]);

  // One entry per logical line of the config, independent of the (debounced) render - typing a
  // newline should renumber the gutter immediately, not on the caller's own debounce timer. Safe
  // to derive straight from `value` because .noWrap forces one visual row per logical line, so
  // "line N" always means the same thing here as it does in the highlighted code below.
  const lineCount = useMemo(() => value.split('\n').length, [value]);
  const lineNumbers = useMemo(
    () => Array.from({ length: lineCount }, (_, i) => i + 1),
    [lineCount],
  );

  const closeCompletion = useCallback(() => {
    setCompletionState({
      open: false,
      items: [],
      selectedIndex: 0,
      top: 0,
      left: 0,
      bottom: 0,
      openUpward: false,
      maxHeight: null,
      replaceStart: 0,
      insideOpenString: false,
    });
    setTooltip((current) => (current.open ? { ...current, open: false } : current));
  }, []);

  // Shared by the explicit Ctrl+Space trigger and the auto-trigger-after-`"` below. Takes the
  // text/cursor explicitly (rather than reading the `value` prop) so the auto-trigger path can
  // pass the textarea's own current DOM value - it fires from a rAF callback where the `value`
  // prop closed over by this callback's own render may not yet include the just-typed character.
  const triggerCompletionFor = useCallback(
    (text, cursorOffset, textarea) => {
      const suggestions = getCompletions(text, format, cursorOffset, schemaScope);
      if (!suggestions.length) {
        closeCompletion();
        return;
      }

      const { start: replaceStart, insideOpenString } = getCompletionReplacement(text, format, cursorOffset, schemaScope);

      const beforeCursor = text.slice(0, cursorOffset);
      const lines = beforeCursor.split('\n');
      const lineText = lines[lines.length - 1] || '';
      const lineHeight = parseFloat(getComputedStyle(textarea).lineHeight) || 24;
      const belowTop = lines.length * lineHeight + lineHeight + 8;
      // Mirrors belowTop but anchored to the popup's BOTTOM edge sitting just above the
      // caret's own line, for when there isn't enough room below (see the flip decision).
      const aboveBottomFromTop = (lines.length - 1) * lineHeight - 8;
      const left = Math.min(Math.max(8, lineText.length * 8), Math.max(8, textarea.clientWidth - 220));

      // Popular editors (VS Code, Monaco) flip a completion popup above the caret when the
      // viewport doesn't have room below, rather than letting it run off-screen or forcing
      // the page to scroll - the existing max-height + overflow-y:auto (styles.module.css)
      // already handles "too many items" in either direction, this only decides which way
      // the popup opens so it stays fully visible on tall AND short viewports alike.
      const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16;
      const popupMaxHeight = 16 * rootFontSize; // keep in sync with .completionPopup's max-height
      const editorBodyRect = editorBodyRef.current?.getBoundingClientRect();
      let openUpward = false;
      let maxHeight = popupMaxHeight;
      let bottom = 0;
      if (editorBodyRect) {
        const spaceBelow = window.innerHeight - (editorBodyRect.top + belowTop);
        const spaceAbove = editorBodyRect.top + aboveBottomFromTop;
        openUpward = spaceBelow < popupMaxHeight && spaceAbove > spaceBelow;
        const available = openUpward ? spaceAbove : spaceBelow;
        // Always leave a small margin off the viewport edge even in the rare case a line
        // sits so close to the very top/bottom that neither direction has full room.
        maxHeight = Math.max(80, Math.min(popupMaxHeight, available - 16));
        bottom = editorBodyRef.current.clientHeight - aboveBottomFromTop;
      }

      setCompletionState({
        open: true,
        items: suggestions,
        selectedIndex: 0,
        top: belowTop,
        left,
        bottom,
        openUpward,
        maxHeight,
        replaceStart,
        insideOpenString,
      });
    },
    [closeCompletion, format, schemaScope],
  );

  const openCompletion = useCallback(
    (event) => {
      const textarea = document.getElementById(textareaId);
      if (!textarea) {
        return;
      }

      triggerCompletionFor(value, textarea.selectionStart, textarea);
      event.preventDefault();
    },
    [textareaId, triggerCompletionFor, value],
  );

  // Auto-trigger right after the user types an opening `"` - the character hasn't landed in the
  // DOM/React state yet during keydown, so this reads the textarea's own value on the next frame
  // instead of relying on the (still stale) `value` prop closed over here.
  const triggerCompletionAfterQuote = useCallback(() => {
    requestAnimationFrame(() => {
      const textarea = document.getElementById(textareaId);
      if (!textarea) {
        return;
      }

      triggerCompletionFor(textarea.value, textarea.selectionStart, textarea);
    });
  }, [textareaId, triggerCompletionFor]);

  const applyCompletion = useCallback(
    (suggestion) => {
      const textarea = document.getElementById(textareaId);
      if (!textarea) {
        return;
      }

      // selectionEnd, not selectionStart: most acceptances are a plain caret (both equal), but
      // the boolean/chainDefault branches below deliberately leave a genuine RANGE selection
      // over the seeded literal (so the reader can just overtype it) - selectionStart would
      // then point at the START of that literal, leaving it stuck in afterCursor instead of
      // being replaced by whatever's accepted from a chained popup.
      const cursorOffset = textarea.selectionEnd;
      const { replaceStart, insideOpenString } = completionState;
      const beforeToken = value.slice(0, replaceStart);
      const afterCursor = value.slice(cursorOffset);

      // The indentation of the line the completion is happening on, reused below to keep
      // freshly-inserted structure (a new sibling line after `,`, a chained object's body)
      // aligned with the rest of the document instead of collapsing everything onto one
      // line - matching the 2-space-indented style every bundled/starter config uses.
      const lineStart = beforeToken.lastIndexOf('\n') + 1;
      const indent = beforeToken.slice(lineStart).match(/^[ \t]*/)[0];

      // Whether a closing quote is still needed: none at all if the item doesn't need
      // quoting (e.g. a bare number/boolean value), the opening one only if the cursor
      // isn't already inside a hand-typed `"`. A `"` sitting right after the cursor is
      // only OUR string's closer if what follows it looks like the right delimiter for
      // this kind (`:` after a key, `,`/`}`/`]`/end after a value) - otherwise it's an
      // unrelated adjacent token (e.g. the next sibling key's own opening quote, pushed
      // right up against our cursor because we're inserting before existing content),
      // and must not be mistaken for a closer we can reuse.
      const afterQuoteRest = afterCursor.startsWith('"') ? afterCursor.slice(1).trimStart() : null;
      const hasQuoteAfter =
        insideOpenString &&
        afterQuoteRest !== null &&
        (suggestion.kind === 'property'
          ? afterQuoteRest.startsWith(':')
          : afterQuoteRest === '' || /^[,}\]]/.test(afterQuoteRest));
      const openingQuote = suggestion.needsQuotes && !insideOpenString ? '"' : '';
      const closingQuote = suggestion.needsQuotes && !hasQuoteAfter ? '"' : '';
      const remainingAfter = hasQuoteAfter ? afterCursor.slice(1) : afterCursor;

      // Auto-complete the punctuation that follows so the user can keep typing without
      // reaching for it themselves - `: ` after a key (if not already there), a comma
      // after a value (ditto). Skipped entirely for a value that's already the last thing
      // in its container (nothing but `}`/`]`/end-of-input follows): a trailing comma there
      // would make the JSON invalid. Otherwise, whether a fresh `\n<indent>` is added too
      // depends on what's already there: appending one unconditionally would double up with
      // an existing sibling that's already on its own line (inserting a property in the
      // middle of an object, say), leaving a blank line between the comma and it.
      const remainingTrimmed = remainingAfter.trimStart();
      const isLastInContainer =
        remainingTrimmed === '' || /^[}\]]/.test(remainingTrimmed);
      const alreadyOnOwnLine = /^[ \t]*\r?\n/.test(remainingAfter);
      // Shared by every chained value below (object/array bodies, a chainDefault literal):
      // whatever the chain inserts still needs the same separator a plain string/number
      // value gets once it's typed, for exactly the same reason.
      const chainSeparator =
        remainingTrimmed.startsWith(',') || isLastInContainer
          ? ''
          : alreadyOnOwnLine
            ? ','
            : `,\n${indent}`;
      let trailingPunctuation = '';
      if (suggestion.kind === 'property' && !remainingTrimmed.startsWith(':')) {
        trailingPunctuation = ': ';
      } else if (suggestion.kind === 'value') {
        trailingPunctuation = chainSeparator;
      }

      const insertText = `${openingQuote}${suggestion.insertText}${closingQuote}${trailingPunctuation}`;

      // Chain straight into the next completion cycle instead of waiting for the user to
      // type the opening character themselves - but only right after we've created a
      // brand-new "key": gap ourselves (trailingPunctuation === ': ') with nothing already
      // sitting in the value slot. What we chain into depends on the property's value type:
      // - string-like: open the quote and immediately trigger the value popup, exactly as
      //   if the user had typed `"` themselves.
      // - object: insert an indented, multi-line empty body (matching how every nested
      //   object in the codebase's own configs is formatted) so the JSON stays valid even
      //   if the user stops right here, cursor placed on the blank inner line, and trigger
      //   its own nested properties.
      // - array: seed a first element matching the array's own item type instead of a bare
      //   `[]` - a quoted empty string (cursor between the quotes, value popup triggered)
      //   for arrays of strings, an indented empty object for arrays of objects, or just the
      //   flat empty container when the item type isn't string/object-shaped.
      // - boolean/integer/number with a schema default: no popup to chain into (there's no
      //   enum), so seed the schema's own default literal instead of leaving the value slot
      //   empty/invalid, with the inserted literal selected so overwriting it (e.g. typing
      //   `true` over a defaulted `false`) is a single keystroke away.
      let chainText = '';
      let chainCursorOffset = 0;
      let chainSelectionStart = 0;
      let shouldTriggerChain = false;
      if (suggestion.kind === 'property' && trailingPunctuation === ': ' && !afterCursor.startsWith('"')) {
        if (suggestion.chainValueType === 'string') {
          // Both quotes and the trailing separator go in up front - same reasoning as the
          // object/array/default branches below: leaving the closing quote for a *later*
          // completion cycle (as a bare opening `"` used to) means the config is invalid
          // JSON the moment the user stops here, and the comma would be missed entirely if
          // they never explicitly complete a value inside it.
          chainText = `""${chainSeparator}`;
          chainCursorOffset = 1;
          chainSelectionStart = 1;
          shouldTriggerChain = true;
        } else if (suggestion.chainValueType === 'object') {
          const innerIndent = `${indent}  `;
          chainText = `{\n${innerIndent}\n${indent}}${chainSeparator}`;
          chainCursorOffset = 2 + innerIndent.length;
          chainSelectionStart = chainCursorOffset;
          shouldTriggerChain = true;
        } else if (suggestion.chainValueType === 'array') {
          if (suggestion.chainItemType === 'string') {
            chainText = `[""]${chainSeparator}`;
            chainCursorOffset = 2;
          } else if (suggestion.chainItemType === 'object') {
            const innerIndent = `${indent}  `;
            chainText = `[{\n${innerIndent}\n${indent}}]${chainSeparator}`;
            chainCursorOffset = 3 + innerIndent.length;
          } else {
            chainText = `[]${chainSeparator}`;
            chainCursorOffset = 1;
          }
          chainSelectionStart = chainCursorOffset;
          shouldTriggerChain = true;
        } else if (suggestion.chainValueType === 'boolean') {
          // Same idea as the string/object/array chains above: seed a real value (the
          // schema's default, or false if it doesn't have one - see getPropertySuggestions)
          // so the config stays valid even if the reader stops right here, but a boolean
          // only ever has two possible values, so it's just as pickable as a small enum -
          // chain into the same true/false popup a "style" or other enum property gets,
          // with the seeded literal highlighted so the reader can also just retype over it.
          const literal = JSON.stringify(suggestion.chainDefault);
          chainText = `${literal}${chainSeparator}`;
          chainCursorOffset = literal.length;
          chainSelectionStart = 0;
          shouldTriggerChain = true;
        } else if (suggestion.chainDefault !== null) {
          // No further popup follows a bare literal, so this is the only chance to add the
          // same trailing separator a chained string/object/array value gets once ITS own
          // completion is accepted - otherwise the config is left syntactically invalid
          // (missing comma) the moment the user accepts the seeded default and moves on.
          const literal = JSON.stringify(suggestion.chainDefault);
          chainText = `${literal}${chainSeparator}`;
          chainCursorOffset = literal.length;
          chainSelectionStart = 0;
        }
      }

      const nextValue = `${beforeToken}${insertText}${chainText}${afterCursor}`;
      onChange(nextValue);
      const chainStartOffset = beforeToken.length + insertText.length;
      const nextSelectionStart = chainStartOffset + chainSelectionStart;
      const nextCursor = chainStartOffset + chainCursorOffset;
      requestAnimationFrame(() => {
        textarea.focus();
        textarea.setSelectionRange(nextSelectionStart, nextCursor);
        if (shouldTriggerChain) {
          triggerCompletionFor(nextValue, nextCursor, textarea);
        }
      });
      closeCompletion();
    },
    [closeCompletion, completionState, textareaId, value, onChange, triggerCompletionFor],
  );

  const handleValueChange = useCallback(
    (nextValue) => {
      onChange(nextValue);

      if (!completionState.open) {
        return;
      }

      // Keep the popup's suggestions in sync as the user keeps typing (e.g. narrowing
      // "type": "se to just "session"), instead of dismissing it on the very next keystroke.
      requestAnimationFrame(() => {
        const textarea = document.getElementById(textareaId);
        if (!textarea) {
          return;
        }

        triggerCompletionFor(textarea.value, textarea.selectionStart, textarea);
      });
    },
    [completionState.open, onChange, textareaId, triggerCompletionFor],
  );

  const handleEditorKeyDown = useCallback(
    (event) => {
      if ((event.ctrlKey || event.metaKey) && event.code === 'Space') {
        openCompletion(event);
        return;
      }

      // Typing an opening quote is the moment a key name or a value is about to start, so it's
      // the most useful point to show suggestions without waiting for an explicit Ctrl+Space.
      // Don't preventDefault - the quote itself must still be typed normally.
      if (event.key === '"' && !event.ctrlKey && !event.metaKey && !event.altKey) {
        triggerCompletionAfterQuote();
      }

      if (!completionState.open) {
        return;
      }

      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault();
          setCompletionState((state) => ({
            ...state,
            selectedIndex: (state.selectedIndex + 1) % state.items.length,
          }));
          break;
        case 'ArrowUp':
          event.preventDefault();
          setCompletionState((state) => ({
            ...state,
            selectedIndex: (state.selectedIndex - 1 + state.items.length) % state.items.length,
          }));
          break;
        case 'Enter':
        case 'Tab':
          event.preventDefault();
          if (completionState.items[completionState.selectedIndex]) {
            applyCompletion(completionState.items[completionState.selectedIndex]);
          }
          break;
        case 'Escape':
          event.preventDefault();
          closeCompletion();
          break;
        default:
          break;
      }
    },
    [
      applyCompletion,
      closeCompletion,
      completionState.items,
      completionState.open,
      completionState.selectedIndex,
      openCompletion,
      triggerCompletionAfterQuote,
    ],
  );

  useEffect(() => {
    if (!completionState.open) {
      return undefined;
    }

    const handleOutsideClick = (event) => {
      if (event.target instanceof Node && !event.target.closest(`.${styles.editorBody}`)) {
        closeCompletion();
      }
    };

    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [closeCompletion, completionState.open]);

  // Shows a richer description next to whichever option the pointer is over - `detail` (the
  // short title) is already visible in the row itself, so this only has something to show when
  // the item actually carries the separate `description` field (see completion.js's
  // normalizeCompletionItem). Positioned relative to editorBody, the same anchor the completion
  // popup itself uses, so it stays correct regardless of where the popup opened.
  const handleOptionMouseEnter = useCallback((index) => {
    setCompletionState((state) => (state.selectedIndex === index ? state : { ...state, selectedIndex: index }));
  }, []);

  // Same tooltip as the completion popup's own, but for hovering a key already typed into the config
  // itself (not just an entry in the completion popup) - lets a reader mouse over any key in
  // a loaded config and see what it does, without having to retrigger completion. Skipped
  // while the popup is open since the two never overlap spatially (the popup covers the area
  // a reader would otherwise be hovering) and would otherwise fight over the same tooltip state.
  const handleEditorMouseMove = useCallback(
    (event) => {
      if (completionState.open) {
        return;
      }

      const textarea = document.getElementById(textareaId);
      const editorBodyRect = editorBodyRef.current?.getBoundingClientRect();
      if (!textarea || !editorBodyRect) {
        return;
      }

      const rect = textarea.getBoundingClientRect();
      const cs = getComputedStyle(textarea);
      const paddingTop = parseFloat(cs.paddingTop) || 0;
      const paddingLeft = parseFloat(cs.paddingLeft) || 0;
      const lineHeight = parseFloat(cs.lineHeight) || 24;
      const charWidth = getCharWidth(textarea) || 8;

      const relY = event.clientY - rect.top - paddingTop + textarea.scrollTop;
      const relX = event.clientX - rect.left - paddingLeft + textarea.scrollLeft;

      const lines = value.split('\n');
      const row = Math.floor(relY / lineHeight);
      const col = Math.round(relX / charWidth);

      if (relY < 0 || relX < 0 || row < 0 || row >= lines.length || col > lines[row].length) {
        setTooltip((current) => (current.open ? { ...current, open: false } : current));
        return;
      }

      const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16;
      const tooltipWidth = 18 * rootFontSize;
      const gap = 12;
      const openLeft = event.clientX + gap + tooltipWidth > window.innerWidth;
      const left = (openLeft ? event.clientX - gap - tooltipWidth : event.clientX + gap) - editorBodyRect.left;
      const maxTop = window.innerHeight - editorBodyRect.top - 4 * rootFontSize;
      const top = Math.min(event.clientY - editorBodyRect.top, Math.max(0, maxTop));

      // The squiggle itself (errorLocation's [column, endColumn) range on its own line) takes
      // priority over a plain key/value hover - it is drawn right there for exactly this reason,
      // and a reader parked on it wants to know what's wrong, not what the token means.
      if (
        errorMessage &&
        errorLocation &&
        row === errorLocation.line - 1 &&
        col >= errorLocation.column - 1 &&
        col < errorLocation.endColumn
      ) {
        setTooltip({
          open: true,
          title: 'Config error',
          text: errorMessage,
          examples: null,
          top,
          left,
        });
        return;
      }

      let offset = 0;
      for (let i = 0; i < row; i += 1) {
        offset += lines[i].length + 1;
      }
      offset += col;

      const info = getHoverInfo(value, format, offset, schemaScope);
      if (!info) {
        setTooltip((current) => (current.open ? { ...current, open: false } : current));
        return;
      }

      setTooltip({
        open: true,
        title: info.title,
        text: info.text,
        examples: info.examples,
        top,
        left,
      });
    },
    [completionState.open, errorLocation, errorMessage, format, schemaScope, textareaId, value],
  );

  const handleEditorMouseLeave = useCallback(() => {
    setTooltip((current) => (current.open ? { ...current, open: false } : current));
  }, []);

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
  //
  // errorLocation's squiggle is drawn here, in the highlight layer, rather than as a separately
  // positioned overlay: this <pre> already scrolls in perfect sync with the textarea on top of
  // it (it's what the textarea is transparent over), so a token wrapped in .errorSquiggle scrolls
  // with the code for free, with none of the manual scroll-offset math a floating overlay would
  // need. A token whose range overlaps the error's [column, endColumn) is split into up to three
  // parts so only the offending slice gets the squiggle, not the whole token.
  const highlightConfig = useCallback(
    (code) => (
      <Highlight prism={Prism} language={format} code={code} theme={codeTheme}>
        {({ tokens, getLineProps, getTokenProps }) => (
          <>
            {tokens.map((line, i) => {
              const onErrorLine = errorLocation && i === errorLocation.line - 1;
              const rangeStart = onErrorLine ? errorLocation.column - 1 : null;
              const rangeEnd = onErrorLine ? errorLocation.endColumn - 1 : null;

              let consumed = 0;

              return (
                <div key={i} {...getLineProps({ line })}>
                  {line.map((token, key) => {
                    const tokenProps = getTokenProps({ token });
                    const start = consumed;
                    const end = start + token.content.length;
                    consumed = end;

                    if (!onErrorLine || end <= rangeStart || start >= rangeEnd) {
                      return <span key={key} {...tokenProps} />;
                    }

                    const before = token.content.slice(0, Math.max(0, rangeStart - start));
                    const mid = token.content.slice(
                      Math.max(0, rangeStart - start),
                      Math.min(token.content.length, rangeEnd - start),
                    );
                    const after = token.content.slice(Math.min(token.content.length, rangeEnd - start));

                    return (
                      <span key={key} {...tokenProps}>
                        {before}
                        <span className={styles.errorSquiggle}>{mid}</span>
                        {after}
                      </span>
                    );
                  })}
                </div>
              );
            })}
          </>
        )}
      </Highlight>
    ),
    [format, codeTheme, errorLocation],
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
        <div ref={editorBodyRef} className={styles.editorBody}>
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
            textareaClassName={classnames(styles.noWrap, styles.editorTextarea)}
            style={{
              overflow: 'auto',
              backgroundColor: codeTheme.plain.backgroundColor,
              color: codeTheme.plain.color,
            }}
            textareaId={textareaId}
            value={value}
            onValueChange={handleValueChange}
            highlight={highlightConfig}
            // The padding has to come through this prop rather than .editor's CSS.
            // react-simple-code-editor absolutely positions its own <pre> and <textarea>
            // to fill the container, and an absolutely positioned box anchors to the
            // padding box - so CSS padding on .editor moved nothing, leaving the code
            // flush against the gutter's border while the gutter, an ordinary flex
            // sibling with the same padding, sat a line lower. This prop is applied
            // inline to both layers, so the code and the numbers share one origin.
            padding={EDITOR_PADDING}
            onKeyDown={handleEditorKeyDown}
            onMouseMove={handleEditorMouseMove}
            onMouseLeave={handleEditorMouseLeave}
          />
          {completionState.open && (
           <div
             className={styles.completionPopup}
             role="listbox"
             aria-label="Config completions"
             style={
               completionState.openUpward
                 ? { bottom: completionState.bottom, left: completionState.left, maxHeight: completionState.maxHeight }
                 : { top: completionState.top, left: completionState.left, maxHeight: completionState.maxHeight }
             }
           >
             {completionState.items.map((item, index) => {
               const isActive = index === completionState.selectedIndex;
               return (
                 <div
                   key={item.label}
                   ref={isActive ? activeCompletionRef : null}
                   className={classnames(styles.completionOption, {
                     [styles.completionOptionActive]: isActive,
                   })}
                   role="option"
                   aria-selected={isActive}
                   onMouseDown={(event) => {
                     event.preventDefault();
                     applyCompletion(item);
                   }}
                   onMouseEnter={() => handleOptionMouseEnter(index)}
                 >
                   <span className={styles.completionLabel}>{item.label}</span>
                   {item.detail && <span className={styles.completionDetail}>{item.detail}</span>}
                 </div>
               );
             })}
           </div>
          )}
          {tooltip.open && (
            <div className={styles.completionTooltip} style={{ top: tooltip.top, left: tooltip.left }} role="tooltip">
              <div className={styles.completionTooltipTitle}>{tooltip.title}</div>
              {tooltip.text && <div className={styles.completionTooltipText}>{tooltip.text}</div>}
              {tooltip.examples && tooltip.examples.length > 0 && (
                <div className={styles.completionTooltipExamples}>
                  Examples: {tooltip.examples.map((value) => (typeof value === 'string' ? value : JSON.stringify(value))).join(', ')}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default ConfigEditor;
