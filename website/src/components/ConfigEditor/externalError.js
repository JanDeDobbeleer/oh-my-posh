// Draws the caller-supplied syntax-error squiggle (errorLocation/errorMessage - see index.js's
// own prop doc comments) as a decoration, entirely separate from @codemirror/lint's diagnostic
// state (there is no schema-validation linter wired in at all - see editorExtensions.js - so
// today that state is simply empty, but the separation still matters): errorLocation is derived
// by the CALLER (Studio/index.js, Config.js) from a client-side re-parse that can catch a
// slightly different, more specific error than whatever CodeMirror-native diagnostics a future
// linter might add - piggybacking on the lint StateField would mean whichever ran last silently
// overwrote the other instead of both being able to show what they know.
import { StateEffect, StateField } from '@codemirror/state';
import { Decoration, EditorView, hoverTooltip } from '@codemirror/view';

const setExternalErrorEffect = StateEffect.define();

// Turns a { line, column, endColumn } (all 1-based, matching errorPosition.js's own convention)
// into a document offset range, or null if the location no longer exists in the current
// document - the error came from a re-parse of a PREVIOUS value of the text (see the callers'
// own debounce), so by the time this runs the doc may already have grown/shrunk past it.
function toDocRange(doc, location) {
  if (!location || location.line < 1 || location.line > doc.lines) {
    return null;
  }

  const line = doc.line(location.line);
  const from = Math.min(line.to, Math.max(line.from, line.from + location.column - 1));
  const to = Math.min(line.to, Math.max(from, line.from + location.endColumn - 1));

  if (to <= from) {
    return null;
  }

  return { from, to, message: location.message };
}

const errorMarkDeco = Decoration.mark({ class: 'cm-omp-external-error' });

// Stores the raw { line, column, endColumn, message } location the caller last supplied, not a
// pre-computed doc range - a plain keystroke elsewhere in the document changes what offset "line
// N" maps to without changing the location itself, so the range is re-derived from the CURRENT
// doc on every read (see the decorations/hover providers below) instead of tracked incrementally.
const externalErrorField = StateField.define({
  create() {
    return null;
  },
  update(value, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setExternalErrorEffect)) {
        value = effect.value;
      }
    }
    return value;
  },
  provide: (field) =>
    EditorView.decorations.of((view) => {
      const range = toDocRange(view.state.doc, view.state.field(field));
      return range ? Decoration.set([errorMarkDeco.range(range.from, range.to)]) : Decoration.none;
    }),
});

// Mirrors ErrorIndicator.js's own tooltip title/copy ("Config error" + the raw message) so
// hovering the squiggle and hovering the warning triangle in the actions row read as the same
// message surfaced two ways, not two different explanations of the same problem.
const externalErrorHover = hoverTooltip((view, pos) => {
  const range = toDocRange(view.state.doc, view.state.field(externalErrorField, false));
  if (!range || pos < range.from || pos > range.to) {
    return null;
  }

  return {
    pos: range.from,
    end: range.to,
    above: true,
    create() {
      const dom = document.createElement('div');
      dom.className = 'cm-omp-external-error-tooltip';
      const title = document.createElement('div');
      title.className = 'cm-omp-external-error-tooltip-title';
      title.textContent = 'Config error';
      const text = document.createElement('div');
      text.textContent = range.message;
      dom.append(title, text);
      return { dom };
    },
  };
});

// baseTheme rather than a styles.module.css rule: this module has no CSS Module of its own, and
// a plain global class here would leak into every consumer of the site's stylesheet - baseTheme
// scopes it to CodeMirror's own generated stylesheet the same way the rest of the theme
// extensions (editorTheme.js) do.
const externalErrorBaseTheme = EditorView.baseTheme({
  '.cm-omp-external-error': {
    textDecoration: 'underline wavy var(--omp-error-color, #f07178)',
    textUnderlineOffset: '2px',
  },
  '.cm-omp-external-error-tooltip': {
    padding: '0.5rem 0.65rem',
    maxWidth: '18rem',
    fontSize: '0.85rem',
    lineHeight: '1.4',
  },
  '.cm-omp-external-error-tooltip-title': {
    fontWeight: '600',
    marginBottom: '0.25rem',
  },
});

export const externalErrorExtension = [externalErrorField, externalErrorHover, externalErrorBaseTheme];

// Dispatched from index.js whenever the errorLocation/errorMessage props change - `error` is
// either { line, column, endColumn, message } or null/undefined for "clear the squiggle".
export function setExternalError(view, error) {
  view.dispatch({ effects: setExternalErrorEffect.of(error || null) });
}
