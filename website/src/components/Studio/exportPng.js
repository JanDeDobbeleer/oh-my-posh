// Rasterizes the studio's live SVG preview into a downloadable PNG.
//
// The SVG's <text> elements render in "Victor Mono" purely by inheriting the *page's* own
// @font-face (see src/css/custom.css / font-metrics.mjs's FONT_FAMILY doc comment) - the SVG
// markup itself carries no font. That is fine for dangerouslySetInnerHTML, which paints inside
// the document, but an <img>/Image() loaded from a blob: URL renders in a sandboxed context with
// no access to the page's stylesheets: without a font of its own, every character cell would
// fall back to a generic monospace font and drift off the grid the SVG's textLength/lengthAdjust
// attributes assume. The fix - the same one every "SVG to PNG" tool uses - is to embed the font
// *inside* the SVG as a base64 data URI before rasterizing it, so the image is fully
// self-contained.
const FONT_URL = '/fonts/VictorMono.ttf';

// Fetched at most once per page load and cached: the font is already sitting in the browser's
// HTTP cache (the page itself loaded it for the on-screen preview), so this only costs a
// base64-encoding pass, not a second network round trip.
let fontDataUriPromise = null;

function getFontDataUri() {
  if (!fontDataUriPromise) {
    fontDataUriPromise = fetch(FONT_URL)
      .then((res) => res.blob())
      .then(
        (blob) =>
          new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(reader.result);
            reader.onerror = () => reject(reader.error);
            reader.readAsDataURL(blob);
          }),
      );
  }

  return fontDataUriPromise;
}

// The SVG's own width/height (svg.go's newCanvasSize) are sized for a 1x screen; rasterizing at
// that size would look soft once downloaded and viewed at natural size or shared/embedded
// elsewhere. 2x matches a typical retina screenshot without ballooning the file for what is
// usually a few short lines of text.
const EXPORT_SCALE = 2;

const SVG_DIMENSIONS_RE = /<svg[^>]*\swidth="([\d.]+)"[^>]*\sheight="([\d.]+)"/;

// svgText is the exact markup already on screen (the same string useWasmRenderer handed to
// dangerouslySetInnerHTML) - this re-encodes it for export rather than re-rendering anything.
export async function exportSvgAsPng(svgText, filename = 'oh-my-posh-prompt.png') {
  const fontDataUri = await getFontDataUri();

  // Inserted as the very first child of <svg>, right after the opening tag, so the font-face
  // exists before the browser lays out any <text> - a plain string splice is enough here since
  // the target is exactly one well-known opening tag, not arbitrary markup.
  const withFont = svgText.replace(
    /(<svg[^>]*>)/,
    `$1<style>@font-face { font-family: "Victor Mono"; src: url(${fontDataUri}) format("truetype"); }</style>`,
  );

  const dimensions = svgText.match(SVG_DIMENSIONS_RE);
  const width = dimensions ? parseFloat(dimensions[1]) : null;
  const height = dimensions ? parseFloat(dimensions[2]) : null;

  const svgBlob = new Blob([withFont], { type: 'image/svg+xml;charset=utf-8' });
  const svgUrl = URL.createObjectURL(svgBlob);

  try {
    const image = new Image();

    await new Promise((resolve, reject) => {
      image.onload = resolve;
      image.onerror = () => reject(new Error('Failed to load the rendered prompt as an image.'));
      image.src = svgUrl;
    });

    // decode() waits for the image's pixels to actually be ready to paint. onload alone fires
    // once the SVG document has parsed, which is not the same guarantee for a data-URI
    // @font-face's glyphs - without this, a fast machine could rasterize the very first frame
    // with a fallback font while the embedded one was still being parsed.
    if (image.decode) {
      await image.decode();
    }

    const canvas = document.createElement('canvas');
    canvas.width = (width ?? image.naturalWidth) * EXPORT_SCALE;
    canvas.height = (height ?? image.naturalHeight) * EXPORT_SCALE;

    const ctx = canvas.getContext('2d');
    ctx.drawImage(image, 0, 0, canvas.width, canvas.height);

    const pngBlob = await new Promise((resolve, reject) => {
      canvas.toBlob(
        (blob) => (blob ? resolve(blob) : reject(new Error('Failed to encode the PNG.'))),
        'image/png',
      );
    });

    const pngUrl = URL.createObjectURL(pngBlob);

    try {
      const link = document.createElement('a');
      link.href = pngUrl;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
    } finally {
      URL.revokeObjectURL(pngUrl);
    }
  } finally {
    URL.revokeObjectURL(svgUrl);
  }
}
