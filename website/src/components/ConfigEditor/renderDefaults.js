// Shared inputs to the wasm render() call (see wasmLoader.js / src/wasm/main.go), factored out
// so every consumer - the studio and, now, each segment doc's own editor (see Config.js) - feeds
// the renderer the exact same sample machine and font metrics rather than each hand-rolling its
// own and quietly drifting apart.
import { VICTOR_MONO } from '../../../font-metrics.mjs';
import segmentData from '../../../segment_data.json';

// Same synthetic data the theme gallery and the build-time segment previews render with (see
// website/export_themes.mjs and scripts/render-segment-previews.mjs) - passed as-is so every
// live preview matches its own build-time/gallery counterpart instead of showing differently
// populated segments.
export const RENDER_DATA_JSON = JSON.stringify(segmentData);

// The four font metrics are Victor Mono's own, read once from font-metrics.mjs so no consumer
// can drift from what the gallery or the build-time preview script pass.
// `columns` is the one dimension callers disagree on: the studio renders a full prompt at the
// gallery's 120, while a single segment (Config.js) renders at the same 64 the build-time
// preview script uses (see scripts/render-segment-previews.mjs's PREVIEW_COLUMNS) so a page's
// live preview lines up with the static SVG it replaces.
export function buildRenderOptions(columns) {
  return {
    columns,
    fontFamily: VICTOR_MONO.FONT_FAMILY,
    cellWidth: VICTOR_MONO.CELL_WIDTH,
    lineHeight: VICTOR_MONO.LINE_HEIGHT,
    fillAscent: VICTOR_MONO.FILL_ASCENT,
    fillDescent: VICTOR_MONO.FILL_DESCENT,
  };
}
