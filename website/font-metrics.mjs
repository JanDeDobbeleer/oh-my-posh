// Victor Mono's own font metrics: the family name plus the four numbers a renderer needs to lay
// out its grid for a font that has no built-in profile. This is the single source of truth for
// every consumer that needs it: website/export_themes.mjs (renders every theme in this font for
// the gallery), scripts/render-segment-previews.mjs (renders each segment doc's own preview at
// build time), and the studio component (website/src/components/Studio) that renders live in the
// browser through the wasm module. This
// file has no Node imports on purpose, so it can be imported from browser code as well as from
// plain Node scripts - keep it that way, or the studio's bundle breaks.
//
// FONT_FAMILY matches the family @font-face declares in src/css/custom.css - the whole point of
// inlining is that these SVGs pick that font up for free once embedded in a page that serves it,
// with no embedding/subsetting/data-URI of their own.
//
// The four metrics are read from VictorMonoNerdFontMono-Regular.ttf (the Mono patched variant
// static/fonts/VictorMono.ttf ships - see the note above the @font-face in src/css/custom.css for
// why that variant): the svg package's --cell-width/--line-height defaults are Hack Nerd Font's
// ratios (src/svg/svg.go's defaultCellWidth/defaultLineHeight), correct for the CLI's own default
// font stack but wrong for Victor Mono, which this build actually renders in. Passing these
// explicitly keeps the grid math (column advance, row height) matched to the font the browser will
// actually lay the text out in.
//
// CELL_WIDTH: ASCII 'M' advance 600 / unitsPerEm 1100 = 0.5455em.
// LINE_HEIGHT: hhea (Ascender - Descender + LineGap) 1550 / unitsPerEm 1100 =
// 1.4091em, times 1.2 (the retired PNG renderer's own lineSpacing constant,
// which the svg package's default also carries - see defaultLineHeight) =
// 1.6909em.
export const VICTOR_MONO = {
  FONT_FAMILY: 'Victor Mono',
  CELL_WIDTH: 0.5455,
  LINE_HEIGHT: 1.6909,
  // The box Victor Mono's powerline separators actually ink, read from the
  // glyf entry for U+E0B0 (and shared verbatim by U+E0B2/E0B4/E0B6/E0B8):
  // y -358..1208 at unitsPerEm 1100, so 1208/1100 = 1.0982em above the
  // baseline and 358/1100 = 0.3255em below it. A segment's background is
  // drawn to this box rather than to LineHeight, which is 19% taller - see
  // svg.Options' FillAscent/FillDescent.
  FILL_ASCENT: 1.0982,
  FILL_DESCENT: 0.3255,
};
