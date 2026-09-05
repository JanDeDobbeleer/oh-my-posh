// The editor's colors, kept independent of the language extensions (editorExtensions.js) and
// swapped into index.js's own theme Compartment whenever @docusaurus/theme-common's colorMode
// flips - the same two palettes the docs' fenced code blocks use (docusaurus.config.js's
// prism.theme/darkTheme: Night Owl Light for light, Night Owl for dark), reproduced as a
// CodeMirror HighlightStyle/EditorView.theme pair so a doc's fenced code
// (prism-react-renderer via @theme/CodeBlock) and this editor keep reading as the same
// family even though the two do not share a tokenizer.
import { EditorView } from '@codemirror/view';
import { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { tags } from '@lezer/highlight';

// Night Owl: prism-react-renderer's nightOwl values. Number and boolean differ there
// (#f78c6c vs #ff5874); the editor's single literal role takes the number's color.
const DARK_COLORS = {
  background: '#011627',
  foreground: '#d6deeb',
  keyword: '#7fdbca',
  property: '#80cbc4',
  string: '#addb67',
  literal: '#f78c6c',
  comment: '#637777',
  punctuation: '#c792ea',
};

// Night Owl Light: prism-react-renderer's nightOwlLight values. Number and boolean differ
// there (#aa0982 vs #bc5454); the editor's single literal role takes the number's color.
const LIGHT_COLORS = {
  background: '#fbfbfb',
  foreground: '#403f53',
  keyword: '#0c969b',
  property: '#0c969b',
  string: '#4876d6',
  literal: '#aa0982',
  comment: '#989fb1',
  punctuation: '#994cc3',
};

// Shared between both palettes: only the color values differ, not which tags map to which
// role. YAML keys can tokenize as either a bare propertyName or, depending on the grammar's own
// node naming, a `definition(propertyName)` wrapper - both are mapped so a yaml key gets the
// same color a json key does regardless of which one @codemirror/lang-yaml happens to emit.
function buildHighlightStyle(colors) {
  return HighlightStyle.define([
    { tag: [tags.propertyName, tags.definition(tags.propertyName), tags.attributeName], color: colors.property },
    { tag: tags.string, color: colors.string },
    { tag: [tags.number, tags.bool, tags.null], color: colors.literal },
    { tag: tags.keyword, color: colors.keyword },
    { tag: tags.comment, color: colors.comment, fontStyle: 'italic' },
    { tag: tags.punctuation, color: colors.punctuation },
  ]);
}

const DARK_HIGHLIGHT_STYLE = buildHighlightStyle(DARK_COLORS);
const LIGHT_HIGHLIGHT_STYLE = buildHighlightStyle(LIGHT_COLORS);

// Mirrors EDITOR_PADDING/.gutter's own padding from the removed react-simple-code-editor setup,
// so line 1 still sits flush with the top of the frame instead of gaining a visible gap now that
// CodeMirror owns its own gutter/content layout.
const CONTENT_PADDING = '0.75rem';

function buildThemeExtension(colors, dark) {
  return EditorView.theme(
    {
      '&': {
        backgroundColor: colors.background,
        color: colors.foreground,
      },
      '.cm-content': {
        fontFamily: 'var(--ifm-font-family-monospace)',
        caretColor: colors.foreground,
        padding: CONTENT_PADDING,
      },
      '.cm-cursor, .cm-dropCursor': {
        borderLeftColor: colors.foreground,
      },
      // Matches the old .gutter's own background/border-less look (styles.module.css) - CM's
      // default gutter otherwise draws a visible seam against `.cm-content` above.
      '.cm-gutters': {
        backgroundColor: colors.background,
        color: colors.comment,
        border: 'none',
      },
      '.cm-activeLineGutter, .cm-activeLine': {
        backgroundColor: 'transparent',
      },
      '.cm-tooltip': {
        backgroundColor: 'var(--omp-card-background)',
        border: '1px solid var(--omp-card-border-color)',
        borderRadius: 'var(--omp-card-radius)',
      },
      '.cm-tooltip.cm-tooltip-autocomplete > ul > li[aria-selected]': {
        backgroundColor: 'var(--ifm-color-primary)',
        color: '#fff',
      },
    },
    { dark },
  );
}

const DARK_THEME_EXTENSION = buildThemeExtension(DARK_COLORS, true);
const LIGHT_THEME_EXTENSION = buildThemeExtension(LIGHT_COLORS, false);

// One array per color mode, swapped whole into index.js's theme Compartment - keeping the
// EditorView.theme (chrome colors) and the HighlightStyle (token colors) paired together here
// means a caller only ever has to think about "dark or light", not keep two separate pieces in
// sync by hand.
export function getThemeExtensions(colorMode) {
  return colorMode === 'dark'
    ? [DARK_THEME_EXTENSION, syntaxHighlighting(DARK_HIGHLIGHT_STYLE)]
    : [LIGHT_THEME_EXTENSION, syntaxHighlighting(LIGHT_HIGHLIGHT_STYLE)];
}
